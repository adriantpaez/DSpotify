package kademlia

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/rapidloop/skv"
	"log"
	"net"
	"time"
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
		Bytes:  make([]byte, 2000),
		Obb:    make([]byte, 2000),
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
	Storage *skv.KVStore
}

func NewServer(key Key, ip net.IP, inPort int, outPort int, database string) *Server {
	server := &Server{
		Contact: Contact{
			Id:   key,
			Ip:   ip,
			Port: inPort,
		},
		InPort:  inPort,
		OutPort: outPort,
		Postman: NewPostman(100, ip, outPort),
	}
	server.Buckets = *NewBucketsTable(server)
	storage, err := skv.Open(database)
	if err != nil {
		log.Printf("ERROR %s\n", err.Error())
	}
	server.Storage = storage
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

func (server *Server) joinToNetwork(known *Contact) {
	time.Sleep(2 * time.Second)
	if known == nil {
		fmt.Println("WARNING: Not known contact")
		return
	}
	fmt.Printf("INFO: Joining to network with known contact: %s\n", fmt.Sprint(*known))
	server.Buckets.Update(known)
	server.LookUp(&server.Contact.Id)
}

func (server *Server) Start(known *Contact) {
	log.Printf("Starting DSpotify server\nID: %s\nIP: %s InPort: %d OutPort: %d\n", hex.EncodeToString(server.Contact.Id[:]), server.Contact.Ip.String(), server.InPort, server.OutPort)
	bridge := make(chan *Request)
	go server.Postman.Start()
	go server.start(bridge)
	go server.joinToNetwork(known)
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
	//msg.FuncCode = msg.FuncCode % 4
	var respB []byte
	switch msg.FuncCode {
	case PING:
		//log.Printf("<-- %s:%d PING", msg.Contact.Ip.String(), msg.Contact.Port)
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
	case STORE:
		//log.Printf("<-- %s:%d STORE", msg.Contact.Ip.String(), msg.Contact.Port)
		server.Store(msg.Args)
	case FIND_NODE:
		//log.Printf("<-- %s:%d FIND_NODE", msg.Contact.Ip.String(), msg.Contact.Port)
		resp := server.FindNode(msg.Args)
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
	case FIND_VALUE:
		//log.Printf("<-- %s:%d FIND_VALUE", msg.Contact.Ip.String(), msg.Contact.Port)
		resp := server.FindValue(msg.Args)
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
	case STORE_NETWORK:
		log.Printf("<-- %s:%d STORE_NETWORK", msg.Contact.Ip.String(), msg.Contact.Port)
		server.StoreNetwork(msg.Args)
	case FIND_VALUE_NETWORK:
		log.Printf("<-- %s:%d FIND_VALUE_NETWORK", msg.Contact.Ip.String(), msg.Contact.Port)
		resp := server.FindValueNetwork(msg.Args)
		respB, err := json.Marshal(&resp)
		if err != nil {
			log.Println("ERROR:", err.Error())
			return
		}
		_, _, err = server.Conn.WriteMsgUDP(respB, nil, r.Addr)
		if err != nil {
			log.Println("ERROR:", err.Error())
			return
		}
	default:
		log.Printf("ERROR: Unexpected function code %d\n", msg.FuncCode)
	}
	if msg.SenderType == KADEMLIA_NODE && msg.Contact.Id.Compare(&server.Contact.Id) != 0 {
		server.Buckets.Update(&msg.Contact)
	}
}
