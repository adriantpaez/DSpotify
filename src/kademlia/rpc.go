package kademlia

import (
	"encoding/json"
	"log"
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
	msg := Message{
		Contact:  server.Contact,
		FuncCode: 0,
		Args:     nil,
	}
	msgB, err := json.Marshal(&msg)
	if err != nil {
		return err
	}
	println(string(msgB))
	return nil
}
