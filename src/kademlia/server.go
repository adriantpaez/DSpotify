package kademlia

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/rapidloop/skv"
	"github.com/skycoin/skycoin/src/api"
	"log"
	"net"
	"net/http"
	"net/rpc"
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
	InPort  int
	Storage *skv.KVStore
	Clients *ClientsManager
}

func NewServer(key Key, ip net.IP, inPort int, database string) *Server {
	server := &Server{
		Contact: Contact{
			Id:   key,
			Ip:   ip,
			Port: inPort,
		},
		InPort: inPort,
	}
	server.Buckets = *NewBucketsTable(server)
	storage, err := skv.Open(database)
	if err != nil {
		log.Printf("ERROR %s\n", err.Error())
	}
	server.Storage = storage
	return server
}

func (server *Server) startRPC() {
	rpcServer := new(RPCServer)
	rpc.Register(rpcServer)
	rpc.HandleHTTP()
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Contact.Ip.String(), server.Contact.Port))
	if err != nil {
		log.Fatal("Starting RPC-server -listen error:", err)
	}
	err = http.Serve(ln, nil)
	if err != nil {
		log.Fatal("Starting RPC-server -serve error:", err)
	}
}

func (server *Server) joinToNetwork(known *Contact) {
	time.Sleep(5 * time.Second)
	if known == nil {
		fmt.Println("WARNING: Not known contact")
		return
	}
	fmt.Printf("INFO: Joining to network with known contact: %s\n", fmt.Sprint(*known))
	server.Buckets.Update(known)
	server.LookUp(&server.Contact.Id)
}

func (server *Server) Start(known *Contact, trackerIp *net.IP, trackerPort int) {
	log.Printf("Starting DSpotify server\nID: %s\nIP: %s InPort: %d \n", hex.EncodeToString(server.Contact.Id[:]), server.Contact.Ip.String(), server.InPort)
	InitClientsManager()
	server.Clients = &clientsManager
	go server.joinToNetwork(known)
	go registerNode(&server.Contact, trackerIp, trackerPort)
	server.startRPC()

}

func registerNode(server *Contact, ip *net.IP, port int) bool {
	time.Sleep(5 * time.Second)
	data, err := json.Marshal(server)
	if err != nil {
		return false
	} else {
		_, err := http.Post(fmt.Sprintf("http://%s:%d/nodes", ip.String(), port), api.ContentTypeJSON, bytes.NewBuffer(data))
		if err != nil {
			return false
		}
	}
	return true
}
