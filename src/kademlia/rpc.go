package kademlia

import (
	"encoding/json"
	"log"
	"net"
)

type Message struct {
	Contact  Contact
	FuncCode uint8
	Args     []byte
}

func Decode(b []byte, end int) (msg Message, err error) {
	err = json.Unmarshal(b[:end], &msg)
	return
}

func Encode(msg *Message) (b []byte, err error) {
	return json.Marshal(msg)
}

func (server *Server) Ping() bool {
	return true
}

func Store(c *Contact) {
	log.Printf("<-- %s:%d STORE", c.Ip.String(), c.Port)
}

func FindNode(c *Contact) {
	log.Printf("<-- %s:%d FIND_NODE", c.Ip.String(), c.Port)
}

func FindValue(c *Contact) {
	log.Printf("<-- %s:%d FIND_VALUE", c.Ip.String(), c.Port)
}

func (server *Server) SendPing(c *Contact) error {
	data := Message{
		Contact:  server.Contact,
		FuncCode: 0,
		Args:     nil,
	}
	dataB, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	responseRecipient := make(chan *Request)
	msg := MessageBinary{
		Receiver: net.UDPAddr{
			IP:   c.Ip,
			Port: c.Port,
			Zone: "",
		},
		ResponseRecipient: responseRecipient,
		Data:              dataB,
	}
	server.Postman.Send(&msg)
	resp := <-responseRecipient
	log.Println(resp)
	return nil
}
