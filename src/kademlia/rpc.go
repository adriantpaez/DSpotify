package kademlia

import (
	"encoding/hex"
	"encoding/json"
	"github.com/rapidloop/skv"
	"log"
	"net"
	"sort"
)

type FuncCode int

const (
	PING          = 0
	STORE         = 1
	FIND_NODE     = 2
	FIND_VALUE    = 3
	STORE_NETWORK = 4
)

type Message struct {
	Contact  Contact
	FuncCode FuncCode
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

func (server *Server) StoreNetwork(args []byte) {
	storeArgs := StoreArgs{}
	err := json.Unmarshal(args, &storeArgs)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		return
	}
	log.Printf("STORE_NETWORK %7d bytes with key %s\n", len(storeArgs.Value), hex.EncodeToString(storeArgs.Key[:]))
	nodes := server.LookUp(&storeArgs.Key)
	for _, node := range nodes {
		go SendStore(&server.Contact, node, storeArgs.Key, storeArgs.Value, server.Postman)
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

func SendMessage(from *Contact, c *Contact, funcCode FuncCode, args []byte, waitResponse bool, postman *Postman) (*Request, error) {
	if from == nil {
		*from = Contact{}
	}
	data := Message{
		Contact:  *from,
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
	postman.Send(&msg)
	if waitResponse {
		resp := <-responseRecipient
		return resp, nil
	}
	return nil, nil
}

func SendPing(from *Contact, c *Contact, postman *Postman) bool {
	log.Printf("--> %s:%d PING\n", c.Ip.String(), c.Port)
	resp, err := SendMessage(from, c, PING, nil, true, postman)
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

func SendStore(from *Contact, c *Contact, key Key, value []byte, postman *Postman) {
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
	_, err = SendMessage(from, c, STORE, argsB, false, postman)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	}
}

func SendFindNode(from *Contact, c *Contact, key *Key, postman *Postman) []Contact {
	log.Printf("--> %s:%d FIND_NODE\n", c.Ip.String(), c.Port)
	argsB, err := json.Marshal(key)
	if err != nil {
		log.Println("ERROR:", err.Error())
		return []Contact{}
	}
	resp, err := SendMessage(from, c, FIND_NODE, argsB, true, postman)
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

func SendFindValue(from *Contact, c *Contact, key *Key, postman *Postman) *FindValueResponse {
	log.Printf("--> %s:%d FIND_VALUE Key: %s\n", c.Ip.String(), c.Port, hex.EncodeToString((*key)[:]))
	argsB, err := json.Marshal(key)
	if err != nil {
		log.Println("ERROR:", err.Error())
		return nil
	}
	resp, err := SendMessage(from, c, FIND_VALUE, argsB, true, postman)
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

func SendStoreNetwork(c *Contact, key *Key, value []byte, postman *Postman) {
	log.Printf("--> %s:%d STORE_NETWORK Key: %s Value: %s\n", c.Ip.String(), c.Port, hex.EncodeToString(key[:]), hex.EncodeToString(value))
	args := StoreArgs{
		Key:   *key,
		Value: value,
	}
	argsB, err := json.Marshal(&args)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		return
	}
	_, err = SendMessage(nil, c, STORE_NETWORK, argsB, false, postman)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	}
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
					channels[t-1] <- SendFindNode(&server.Contact, result[k], key, server.Postman)
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
