package kademlia

import (
	"encoding/binary"
	"log"
)

type Message struct {
	Contact  Contact
	FuncCode uint8
	Args     []byte
}

type DecodeError struct {
	Message string
}

func (err DecodeError) Error() string {
	return err.Message
}

func Decode(b []byte, end int) (msg Message, err error) {
	if end < 29 {
		err = DecodeError{Message: "Short message"}
		return
	}
	err = nil
	i := 0
	key := Key{}
	for ; i < KEYSIZE; i++ {
		key[i] = b[i]
	}
	msg.Contact.Id = key
	ip := make([]byte, 4)
	for j := 0; j < 4; j++ {
		ip[j] = b[i]
		i++
	}
	msg.Contact.Ip = ip
	port := int(binary.BigEndian.Uint32(b[i : i+4]))
	msg.Contact.Port = port
	i += 4
	msg.FuncCode = b[i]
	return
}

func (server *Server) Ping(msg *Message) {
	log.Printf("<-- %s:%d PING", msg.Contact.Ip.String(), msg.Contact.Port)
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
