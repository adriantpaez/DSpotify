package kademlia

import (
	"log"
	"net"
	"time"
)

type MessageBinary struct {
	Receiver          net.UDPAddr
	ResponseRecipient chan *Request
	Data              []byte
}

type PostBox struct {
	Id      int
	Message chan *MessageBinary
	Busy    bool
}

func NewBuzon(id int, bufferSize int) *PostBox {
	return &PostBox{
		Id:      id,
		Message: make(chan *MessageBinary, bufferSize),
		Busy:    false,
	}
}

func (b PostBox) Start(port int, onFree chan *PostBox) {
	for {
		msg := <-b.Message
		b.Busy = true
		con, err := net.DialUDP("udp", &net.UDPAddr{
			IP:   nil,
			Port: port,
			Zone: "",
		}, &msg.Receiver)
		if err != nil {
			log.Println("ERROR on PostBox:", err.Error())
			b.send(msg, nil, onFree)
			return
		}
		_, _, err = con.WriteMsgUDP(msg.Data, nil, nil)
		if err != nil {
			log.Println("ERROR on PostBox:", err.Error())
			b.send(msg, nil, onFree)
			return
		}
		err = con.SetReadDeadline(time.Now().Add(2 * 1e9))
		if err != nil {
			log.Println("ERROR on PostBox:", err.Error())
			b.send(msg, nil, onFree)
			return
		}
		r := NewRequest()
		r.NBytes, _, _, r.Addr, r.Err = con.ReadMsgUDP(r.Bytes, r.Obb)
		b.send(msg, r, onFree)
		con.Close()
	}
}

func (b PostBox) send(msg *MessageBinary, r *Request, onFree chan *PostBox) {
	msg.ResponseRecipient <- r
	b.Busy = false
	onFree <- &b
}

type Postman struct {
	Queue     chan *MessageBinary
	Port      int
	PostBoxes []*PostBox
}

func NewPostman(bufferSize int, boxCount int, port int) *Postman {
	p := &Postman{
		Queue:     make(chan *MessageBinary, bufferSize),
		PostBoxes: make([]*PostBox, boxCount),
		Port:      port,
	}
	for i := 0; i < len(p.PostBoxes); i++ {
		p.PostBoxes[i] = NewBuzon(i, bufferSize/boxCount)
	}
	return p
}

func (p Postman) Start() {
	boxMap := map[int]net.UDPAddr{}
	onFree := make(chan *PostBox, len(p.PostBoxes))
	for i := 0; i < len(p.PostBoxes); i++ {
		onFree <- p.PostBoxes[i]
		go p.PostBoxes[i].Start(p.Port, onFree)
	}
	for {
		msg := <-p.Queue
		if p.findBusy(msg, boxMap) {
			continue
		}
		b := <-onFree
		boxMap[b.Id] = msg.Receiver
		b.Message <- msg
	}
}

func (p Postman) Send(msg *MessageBinary) {
	p.Queue <- msg
}

func (p Postman) findBusy(msg *MessageBinary, boxMap map[int]net.UDPAddr) bool {
	for i := 0; i < len(p.PostBoxes); i++ {
		b := p.PostBoxes[i]
		if b.Busy && boxMap[b.Id].IP.Equal(msg.Receiver.IP) && boxMap[b.Id].Port == msg.Receiver.Port {
			p.PostBoxes[i].Message <- msg
			return true
		}
	}
	return false
}
