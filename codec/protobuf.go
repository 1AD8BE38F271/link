package codec

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"reflect"
	"github.com/golang/protobuf/proto"
	"github.com/FTwOoO/link"
)

type protobufPacketHeader struct {
	MessageType uint16
	ContentSize uint16
	Hash        uint64
}

func (d *protobufPacketHeader) HeaderSize() int {
	return binary.Size(d.MessageType) + binary.Size(d.ContentSize) + binary.Size(d.Hash)
}

func (d *protobufPacketHeader) FromBytes(b []byte) (error) {
	if len(b) < d.HeaderSize() {
		return errors.New("Need more data")
	}

	d.MessageType = binary.BigEndian.Uint16(b[:2])
	d.ContentSize = binary.BigEndian.Uint16(b[2:4])
	d.Hash = binary.BigEndian.Uint64(b[4:])
	return nil
}

func (d *protobufPacketHeader) ToBytes() []byte {
	buf := make([]byte, d.HeaderSize())
	binary.BigEndian.PutUint16(buf[:2], d.MessageType)
	binary.BigEndian.PutUint16(buf[2:4], d.ContentSize)
	binary.BigEndian.PutUint64(buf[4:], d.Hash)
	return buf
}

func (d *protobufPacketHeader) ValidateContent(body []byte) error {
	return nil
}

type ProtobufProtocol struct {
	maxRecv        int
	maxSend        int
	valueToMsgType map[uint16]reflect.Type
	msgTypeToValue map[reflect.Type]uint16
	context interface{}
}

func NewProtobufProtocol(context interface{}, msgNames []string) *ProtobufProtocol {
	p := &ProtobufProtocol{}


	p.maxRecv = math.MaxUint16
	p.maxSend =  math.MaxUint16
	p.valueToMsgType = map[uint16]reflect.Type{}
	p.msgTypeToValue = map[reflect.Type]uint16{}
	p.context = context

	for i, msgName := range msgNames {
		t := proto.MessageType(msgName)
		p.valueToMsgType[uint16(i)] = t
		p.msgTypeToValue[t] = uint16(i)
	}
	return p
}

func (d *ProtobufProtocol) EncodeMessage(msg proto.Message) (packet []byte, err error) {
	body, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	h := new(protobufPacketHeader)
	h.MessageType = d.msgTypeToValue[reflect.TypeOf(msg)]
	h.ContentSize = uint16(len(body))
	h.Hash = 0

	if err != nil {
		return
	}

	header := h.ToBytes()
	packet = make([]byte, len(header) + len(body))
	copy(packet, header)
	copy(packet[len(header):], body)
	return
}

func (d *ProtobufProtocol) DecodeHeader(header []byte) (h *protobufPacketHeader, err error) {

	h = new(protobufPacketHeader)
	err = h.FromBytes(header)
	return
}

func (d *ProtobufProtocol) DecodeBody(body []byte, h *protobufPacketHeader) (msg proto.Message, err error) {
	if len(body) != int(h.ContentSize) {
		return nil, errors.New("Content size dont match")
	}

	if err := h.ValidateContent(body); err != nil {
		return nil, err
	}

	if _, ok := d.valueToMsgType[h.MessageType]; !ok {
		return nil, errors.New("Invalid type")
	}

	T := d.valueToMsgType[h.MessageType]
	if T == nil {
		return nil, errors.New("No Message configured for this type")
	}

	msg = reflect.New(T.Elem()).Interface().(proto.Message)
	err = proto.Unmarshal(body, msg)
	if err != nil {
		return nil, err
	}
	return
}

func (p *ProtobufProtocol) NewCodec(rw io.ReadWriter) (cc link.Codec, err error) {
	codec := &protobufCodec{
		rw: rw,
		ProtobufProtocol: p,
		headBuf: make([]byte, new(protobufPacketHeader).HeaderSize()),
	}
	cc = codec
	return
}

type protobufCodec struct {
	headBuf []byte
	bodyBuf []byte
	rw      io.ReadWriter
	*ProtobufProtocol
}

func (c *protobufCodec) Receive() (interface{}, error) {
	if _, err := io.ReadFull(c.rw, c.headBuf); err != nil {
		return nil, err
	}
	header, err := c.DecodeHeader(c.headBuf)
	if err != nil {
		return nil, err

	}
	size := header.ContentSize

	if int(size) > c.maxRecv {
		return nil, errors.New("Too Large Packet")
	}
	if cap(c.bodyBuf) < int(size) {
		c.bodyBuf = make([]byte, size, size + 128)
	}
	buff := c.bodyBuf[:size]
	if _, err := io.ReadFull(c.rw, buff); err != nil {
		return nil, err
	}

	msg, err := c.DecodeBody(buff, header)
	return msg, err
}

func (c *protobufCodec) Send(msg interface{}) error {
	switch msg.(type) {
	case proto.Message:
		all, err := c.EncodeMessage(msg.(proto.Message))
		if err != nil {
			return err
		}

		_, err = c.rw.Write(all)
		if err != nil {
			return err
		}

		return nil
	default:
		return errors.New("Type is not valid")
	}

}

func (c *protobufCodec) Close() error {
	if closer, ok := c.rw.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
