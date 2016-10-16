package codec

import (
	"testing"
	"bytes"
	"reflect"
)


func compareTestPacket(t *testing.T, msg1 *TestPacket, msg2 *TestPacket) {
	// Now test and newTest contain the same data.
	if msg1.Sid != msg2.Sid {
		t.Fatalf("data mismatch %d != %d", msg1.Sid, msg2.Sid)
	}
	if msg1.Mark != msg2.Mark {
		t.Fatalf("data mismatch %q != %q", msg1.Mark, msg2.Mark)
	}

	if !reflect.DeepEqual(msg1.Sessions, msg2.Sessions) {
		t.Fatal("Sessions dont mismatch")
	}
}


func TestProtobufCodec(t *testing.T) {
	var stream bytes.Buffer
	protocol := NewProtobufProtocol([]reflect.Type{reflect.TypeOf(TestPacket{})})

	codec,  _ := protocol.NewCodec(&stream)

	sendMsg1 := &TestPacket{
		Mark: false,
		Sid:  999,
		Sessions: map[string]uint64{"a":1, "b":2},

	}

	sendMsg2 := &TestPacket{
		Mark: true,
		Sid:  18,
		Sessions: map[string]uint64{"xxx":10, "yyy":20},

	}

	err := codec.Send(sendMsg1)
	if err != nil {
		t.Fatal(err)
	}

	err = codec.Send(sendMsg2)
	if err != nil {
		t.Fatal(err)
	}


	recvMsg, err := codec.Receive()
	if err != nil {
		t.Fatal(err)
	}
	recvMsg1 := recvMsg.(*TestPacket)

	recvMsg, err = codec.Receive()
	if err != nil {
		t.Fatal(err)
	}
	recvMsg2 := recvMsg.(*TestPacket)

	compareTestPacket(t, sendMsg1, recvMsg1)
	compareTestPacket(t, sendMsg2, recvMsg2)

}
