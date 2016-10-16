package main

import (
	"log"

	"github.com/FTwOoO/link"
	"github.com/FTwOoO/link/codec"
	"net"
)

type AddReq struct {
	A, B int
}

type AddRsp struct {
	C int
}

func main() {
	json := codec.Json()
	json.Register(AddReq{})
	json.Register(AddRsp{})

	l, err := net.Listen("tcp", "0.0.0.0:0")
	checkErr(err)

	server := link.NewServer(l, json, 0 /* sync send */)
	checkErr(err)
	addr := server.Listener().Addr().String()
	go server.Serve(link.HandlerFunc(serverSessionLoop))


	client := link.NewClient(link.DialerFunc(func()(net.Conn, error){return net.Dial("tcp", addr)}), json, 0)
	checkErr(err)

	session, _ := client.GetSession()
	clientSessionLoop(session)
}

func serverSessionLoop(session *link.Session) {
	for {
		req, err := session.Receive()
		checkErr(err)

		err = session.Send(&AddRsp{
			req.(*AddReq).A + req.(*AddReq).B,
		})
		checkErr(err)
	}
}

func clientSessionLoop(session *link.Session) {
	for i := 0; i < 10; i++ {
		err := session.Send(&AddReq{
			i, i,
		})
		checkErr(err)
		log.Printf("Send: %d + %d", i, i)

		rsp, err := session.Receive()
		checkErr(err)
		log.Printf("Receive: %d", rsp.(*AddRsp).C)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
