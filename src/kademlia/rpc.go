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

func Store(args []byte) {
	log.Println(args)
}

func FindNode(c *Contact) {
	log.Printf("<-- %s:%d FIND_NODE", c.Ip.String(), c.Port)
}

func FindValue(c *Contact) {
	log.Printf("<-- %s:%d FIND_VALUE", c.Ip.String(), c.Port)
}

func (server Server) SendMessage(c *Contact, funcCode uint8, args []byte, waitResponse bool) (*Request, error) {
	data := Message{
		Contact:  server.Contact,
		FuncCode: funcCode,
		Args:     args,
	}
	dataB, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}
	var responseRecipient chan *Request = nil
	if waitResponse {
		responseRecipient = make(chan *Request)
	}
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
	if waitResponse {
		resp := <-responseRecipient
		return resp, nil
	}
	return nil, nil
}

func (server Server) SendPing(c *Contact) bool {
	resp, err := server.SendMessage(c, 0, nil, true)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	} else if resp.Err != nil {
		log.Printf("ERROR: %s\n", resp.Err.Error())
	} else {
		var r bool
		err = json.Unmarshal(resp.Bytes[:resp.NBytes], &r)
		if err != nil {
			log.Printf("ERROR: %s\n", err.Error())
		} else {
			return r
		}
	}
	return false
}

func (server Server) SendStore(c *Contact, value []byte) {
	_, err := server.SendMessage(c, 1, value, false)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	}
}
