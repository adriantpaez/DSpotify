package kademlia

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net"
)

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
		Bytes:  make([]byte, 1000),
		Obb:    make([]byte, 1000),
		Err:    nil,
	}
	return &r
}

type Server struct {
	Contact Contact
	Buckets BucketsTable
	Conn    *net.UDPConn
	InPort  int
	OutPort int
	Postman *Postman
}

func NewServer(key Key, ip net.IP, inPort int, outPort int) *Server {
	server := &Server{
		Contact: Contact{
			Id:   key,
			Ip:   ip,
			Port: inPort,
		},
		InPort:  inPort,
		OutPort: outPort,
		Postman: NewPostman(100, 1, outPort),
	}
	server.Buckets = *NewBucketsTable(server.Contact)
	return server
}

func (server *Server) start(bridge chan *Request) {
	ln, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   server.Contact.Ip,
		Port: server.Contact.Port,
		Zone: "",
	})
	if err != nil {
		log.Println("ERROR:", err.Error())
		return
	}
	server.Conn = ln

	for true {
		r := NewRequest()
		r.NBytes, _, _, r.Addr, r.Err = ln.ReadMsgUDP(r.Bytes, r.Obb)
		bridge <- r
	}
}

func (server *Server) Start() {
	log.Printf("Starting DSpotify server\nID: %s\nIP: %s InPort: %d OutPort: %d\n", hex.EncodeToString(server.Contact.Id[:]), server.Contact.Ip.String(), server.InPort, server.OutPort)
	bridge := make(chan *Request)
	go server.Postman.Start()
	go server.start(bridge)
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
	if !server.Buckets.Update(&msg.Contact) {
		log.Printf("ERROR: Updating buckets table with %s\n", hex.EncodeToString(msg.Contact.Id[:]))
	}
	msg.FuncCode = msg.FuncCode % 4
	var respB []byte
	switch msg.FuncCode {
	case 0:
		log.Printf("<-- %s:%d PING", msg.Contact.Ip.String(), msg.Contact.Port)
		resp := server.Ping()
		respB, err = json.Marshal(&resp)
		if err != nil {
			log.Println("ERROR:", err.Error())
			return
		}
		_, _, err = server.Conn.WriteMsgUDP(respB, nil, r.Addr)
		if err != nil {
			log.Println("ERROR:", err.Error())
			return
		}
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
