package kademlia

import (
	"log"
	"net"
	"time"
)

var mapReady chan int

type MessageBinary struct {
	Receiver          net.UDPAddr
	ResponseRecipient chan *Request
	Data              []byte
}

type PostBox struct {
	Id        string
	Conn      *net.UDPConn
	Message   chan *MessageBinary
	LastWrite time.Time
}

func NewPostBox(conn *net.UDPConn, bufferSize int) *PostBox {
	return &PostBox{
		Id:      conn.RemoteAddr().String(),
		Conn:    conn,
		Message: make(chan *MessageBinary, bufferSize),
	}
}

func (b PostBox) Start(boxMap map[string]*PostBox) {
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case msg := <-b.Message:
			_, _, err := b.Conn.WriteMsgUDP(msg.Data, nil, nil)
			if err != nil {
				log.Println("ERROR on PostBox:", err.Error())
			} else if msg.ResponseRecipient != nil {
				err = b.Conn.SetReadDeadline(time.Now().Add(time.Second * 3))
				if err != nil {
					log.Println("ERROR on PostBox:", err.Error())
				} else {
					r := NewRequest()
					r.NBytes, _, _, r.Addr, r.Err = b.Conn.ReadMsgUDP(r.Bytes, r.Obb)
					msg.ResponseRecipient <- r
				}
			}
			b.LastWrite = time.Now()
		case <-ticker.C:
			<-mapReady
			if b.LastWrite.Add(30 * time.Second).Before(time.Now()) {
				err := b.Conn.Close()
				if err != nil {
					log.Printf("ERROR: %s\n")
				} else {
					delete(boxMap, b.Id)
					ticker.Stop()
					return
				}
			}
			mapReady <- 1
		}
	}
}

type Postman struct {
	Queue  chan *MessageBinary
	IP     net.IP
	Port   int
	BoxMap map[string]*PostBox
}

func NewPostman(bufferSize int, ip net.IP, port int) *Postman {
	p := &Postman{
		Queue:  make(chan *MessageBinary, bufferSize),
		BoxMap: map[string]*PostBox{},
		Port:   port,
		IP:     ip,
	}
	return p
}

func (p Postman) Start() {
	mapReady = make(chan int, 1)
	mapReady <- 1
	for {
		msg := <-p.Queue
		<-mapReady
		if value, ok := p.BoxMap[msg.Receiver.String()]; ok {
			value.Message <- msg
		} else {
			conn, err := net.DialUDP("udp", &net.UDPAddr{
				IP:   p.IP,
				Port: p.Port,
				Zone: "",
			}, &msg.Receiver)
			if err != nil {
				log.Printf("ERROR: %s\n", err.Error())
			} else {
				newBox := NewPostBox(conn, 100)
				newBox.Message <- msg
				p.BoxMap[msg.Receiver.String()] = newBox
				go newBox.Start(p.BoxMap)
			}
		}
		mapReady <- 1
	}
}

func (p Postman) Send(msg *MessageBinary) {
	p.Queue <- msg
}
