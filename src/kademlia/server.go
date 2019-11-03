package kademlia

import (
	"log"
	"net"
)

type Server struct {
	Contact Contact
	Buckets BucketsTable
}

type Request struct {
	Addr   *net.UDPAddr
	NBytes int
	Bytes  []byte
	Obb    []byte
	Err    error
}

func NewRequest() *Request {
	r := Request{
		Addr:   nil,
		NBytes: 0,
		Bytes:  make([]byte, 200),
		Obb:    make([]byte, 200),
		Err:    nil,
	}
	return &r
}

func NewServer(key Key, ip net.IP, port int) *Server {
	server := &Server{Contact: Contact{
		Id:   key,
		Ip:   ip,
		Port: port,
	}}
	server.Buckets = *NewBucketsTable(server.Contact)
	return server
}

func (server *Server) Start(bridge chan *Request) {
	ln, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   server.Contact.Ip,
		Port: server.Contact.Port,
		Zone: "",
	})
	if err != nil {
		print(err.Error())
		return
	}
	for true {
		r := NewRequest()
		r.NBytes, _, _, r.Addr, r.Err = ln.ReadMsgUDP(r.Bytes, r.Obb)
		bridge <- r
	}
}

func (server *Server) Consumer(bridge chan *Request) {
	for true {
		r := <-bridge
		if r.Err == nil {
			go server.handler(r)
		} else {
			log.Println(r.Err.Error())
		}
	}
}

func (server *Server) handler(r *Request) {
	msg, err := Decode(r.Bytes, r.NBytes)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		return
	}
	server.Buckets.Update(&msg.Contact)
	msg.FuncCode = msg.FuncCode % 4
	switch msg.FuncCode {
	case 0:
		server.Ping(&msg)
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
