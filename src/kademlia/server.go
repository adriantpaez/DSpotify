package kademlia

import (
	"log"
	"net"
)

type Server struct {
	Self    Contact
	Buckets BucketsTable
}

func NewServer(key Key, ip net.IP, port int) *Server {
	server := &Server{Self: Contact{
		Id:   key,
		Ip:   ip,
		Port: port,
	}}
	server.Buckets = *NewBucketsTable(server.Self)
	return server
}

func (server *Server) Start() {
	ln, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   server.Self.Ip,
		Port: server.Self.Port,
		Zone: "",
	})
	if err != nil {
		print(err.Error())
		return
	}
	for true {
		addr := &net.UDPAddr{}
		var n int
		b := make([]byte, 200)
		oob := make([]byte, 200)
		n, _, _, addr, err = ln.ReadMsgUDP(b, oob)
		go server.handler(*addr, n, b, err)
	}
}

func (server *Server) handler(addr net.UDPAddr, n int, b []byte, err error) {
	if err != nil {
		log.Println(err.Error())
	} else {
		msg, err := Decode(b, n)
		if err != nil {
			log.Printf("ERROR: %s\n", err.Error())
			return
		}
		server.Buckets.Update(&msg.Contact)
		msg.FuncCode = msg.FuncCode % 4
		switch msg.FuncCode {
		case 0:
			Ping(&msg.Contact)
		case 1:
			Store(&msg.Contact)
		case 2:
			FindNode(&msg.Contact)
		case 3:
			FindValue(&msg.Contact)
		default:
			log.Printf("ERROR: Unexpected function code %d\n", msg.FuncCode)
		}
	}
}
