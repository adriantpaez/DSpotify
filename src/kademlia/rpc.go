package kademlia

import (
	"encoding/hex"
	"encoding/json"
	"github.com/rapidloop/skv"
	"log"
	"net"
	"sort"
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

func (server *Server) Store(args []byte) {
	storeArgs := StoreArgs{}
	err := json.Unmarshal(args, &storeArgs)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		return
	}
	log.Printf("STORE %7d bytes with key %s\n", len(storeArgs.Value), hex.EncodeToString(storeArgs.Key[:]))
	err = server.Storage.Put(hex.EncodeToString(storeArgs.Key[:]), storeArgs.Value)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	}
}

func (server Server) FindNode(args []byte) []Contact {
	var key Key
	err := json.Unmarshal(args, &key)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		return []Contact{}
	}
	kNears := server.Buckets.KNears(&key)
	var resp []Contact
	for _, c := range kNears {
		resp = append(resp, *c)
	}
	return resp
}

type FindValueResponse struct {
	Value  []byte
	KNears []Contact
}

func (server Server) FindValue(args []byte) FindValueResponse {
	resp := FindValueResponse{}
	var k Key
	err := json.Unmarshal(args, &k)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	} else if err := server.Storage.Get(hex.EncodeToString(k[:]), &resp.Value); err == skv.ErrNotFound {
		kNears := server.Buckets.KNears(&k)
		for _, c := range kNears {
			resp.KNears = append(resp.KNears, *c)
		}
	} else if err != nil {
		log.Println("ERROR:", err.Error())
	}
	return resp
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
	log.Printf("--> %s:%d PING\n", c.Ip.String(), c.Port)
	resp, err := server.SendMessage(c, 0, nil, true)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	} else if resp == nil {
		return false
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

type StoreArgs struct {
	Key   Key
	Value []byte
}

func (server Server) SendStore(c *Contact, key Key, value []byte) {
	log.Printf("--> %s:%d STORE Key: %s Value: %s\n", c.Ip.String(), c.Port, hex.EncodeToString(key[:]), hex.EncodeToString(value))
	args := StoreArgs{
		Key:   key,
		Value: value,
	}
	argsB, err := json.Marshal(&args)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		return
	}
	_, err = server.SendMessage(c, 1, argsB, false)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	}
}

func (server Server) SendFindNode(c *Contact, key *Key) []Contact {
	log.Printf("--> %s:%d FIND_NODE\n", c.Ip.String(), c.Port)
	argsB, err := json.Marshal(key)
	if err != nil {
		log.Println("ERROR:", err.Error())
		return []Contact{}
	}
	resp, err := server.SendMessage(c, 2, argsB, true)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	} else if resp == nil {
		return []Contact{}
	} else if resp.Err != nil {
		log.Println("ERROR:", resp.Err.Error())
	} else {
		var r []Contact
		err := json.Unmarshal(resp.Bytes[:resp.NBytes], &r)
		if err != nil {
			log.Println("ERROR:", err.Error())
		} else {
			return r
		}
	}
	return []Contact{}
}

func (server Server) SendFindValue(c *Contact, key *Key) *FindValueResponse {
	log.Printf("--> %s:%d FIND_VALUE Key: %s\n", c.Ip.String(), c.Port, hex.EncodeToString((*key)[:]))
	argsB, err := json.Marshal(key)
	if err != nil {
		log.Println("ERROR:", err.Error())
		return nil
	}
	resp, err := server.SendMessage(c, 3, argsB, true)
	if err != nil {
		log.Println("ERROR:", err.Error())
	} else if resp == nil {
		return nil
	} else if resp.Err != nil {
		log.Println("ERROR:", resp.Err.Error())
	} else {
		var r FindValueResponse
		err := json.Unmarshal(resp.Bytes[:resp.NBytes], &r)
		if err != nil {
			log.Println("ERROR:", err.Error())
		} else {
			return &r
		}
	}
	return nil
}

func (server Server) LookUp(key *Key) []*Contact {
	result := server.Buckets.KNears(key)
	visit := map[Key]bool{}

	for _, c := range result {
		visit[c.Id] = false
	}

	for {
		tmp := 3
		channels := make([]chan []Contact, 3)
		for i := 0; i < len(channels); i++ {
			channels[i] = make(chan []Contact)
		}
		for i := 0; tmp != 0 && i < len(result); i++ {
			if !visit[result[i].Id] {
				go func(t, k int) {
					channels[t-1] <- server.SendFindNode(result[k], key)
				}(tmp, i)
				tmp -= 1
				visit[result[i].Id] = true
			}
		}
		if tmp == 3 {
			return result
		}
		newsIndex := len(result)
		for tmp != 3 {
			select {
			case c := <-channels[0]:
				for _, v := range c {
					result = append(result, &v)
				}
				tmp += 1
			case c := <-channels[1]:
				for _, v := range c {
					result = append(result, &v)
				}
				tmp += 1
			case c := <-channels[2]:
				for _, v := range c {
					result = append(result, &v)
				}
				tmp += 1
			}
		}

		for i := newsIndex; i < len(result); i++ {
			visit[result[i].Id] = false
		}

		sort.Slice(result, func(i, j int) bool {
			distI := result[i].Id.DistanceTo(key)
			distJ := result[j].Id.DistanceTo(key)
			return distI.Compare(&distJ) == -1
		})
	}
}
