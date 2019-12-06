package kademlia

import (
	"encoding/hex"
	"fmt"
	"github.com/rapidloop/skv"
	"log"
	"net/rpc"
	"sort"
	"time"
)

var clientsManager ClientsManager

type ClientState struct {
	Client  *rpc.Client
	LastUse time.Time
}

type ClientsManager struct {
	Clients map[string]*ClientState
	Lock    chan bool
}

func InitClientsManager() {
	clientsManager.Clients = make(map[string]*ClientState)
	clientsManager.Lock = make(chan bool, 1)
	clientsManager.Lock <- true
}

func (m *ClientsManager) GetClient(c *Contact) (*rpc.Client, error) {
	<-m.Lock
	key := fmt.Sprintf("%s:%d", c.Ip.String(), c.Port)
	if client, ok := m.Clients[key]; ok {
		m.Lock <- true
		return client.Client, nil
	}
	client, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%d", c.Ip.String(), c.Port))
	if err == nil {
		m.Clients[key] = &ClientState{
			Client:  client,
			LastUse: time.Now(),
		}
	}
	m.Lock <- true
	return client, err
}

func (m *ClientsManager) Cleaner() {
	for {
		time.Sleep(30 * time.Second)
		<-m.Lock
		var remove []string
		for key, value := range m.Clients {
			if value.LastUse.Add(30 * time.Second).Before(time.Now()) {
				if err := value.Client.Close(); err == nil {
					remove = append(remove, key)
				}
			}
		}
		for _, k := range remove {
			delete(m.Clients, k)
		}
		m.Lock <- true
	}
}

type RPCServer struct{}

type PingArgs struct {
	Contact Contact
}

type PingFromClientArgs struct{}

type StoreArgs struct {
	Contact Contact
	Key     Key
	Value   []byte
}

type FindNodeArgs struct {
	Contact Contact
	Key     Key
}

type FindValueArgs struct {
	Contact Contact
	Key     Key
}

type FindValueResponse struct {
	Value  []byte
	KNears []Contact
}

func (rpc *RPCServer) Ping(args *PingArgs, reply *bool) error {
	server.Buckets.Update(&args.Contact)
	*reply = true
	return nil
}

func (rpc *RPCServer) PingFromClient(args *PingFromClientArgs, reply *bool) error {
	*reply = true
	return nil
}

func (rpc *RPCServer) Store(args *StoreArgs, reply *bool) error {
	server.Buckets.Update(&args.Contact)
	err := server.Storage.Put(hex.EncodeToString(args.Key[:]), args.Value)
	if err != nil {
		*reply = false
		fmt.Println(err.Error())
	}
	*reply = true
	return nil
}

func (rpc *RPCServer) StoreNetwork(args *StoreArgs, reply *bool) error {
	log.Printf("STORE_NETWORK %7d bytes with key %s\n", len(args.Value), hex.EncodeToString(args.Key[:]))
	*reply = false
	nodes := server.LookUp(&args.Key)
	count := len(nodes)
	resp := make(chan bool, count)
	for i := 0; i < len(nodes); i++ {
		go func(i int) {
			resp <- SendStore(&server.Contact, nodes[i], args.Key, args.Value)
		}(i)
	}
	for count != 0 {
		r := <-resp
		if r {
			*reply = true
			break
		}
		count--
	}
	return nil
}

func (rpc *RPCServer) FindNode(args *FindNodeArgs, reply *[]Contact) error {
	server.Buckets.Update(&args.Contact)
	kNears := server.Buckets.KNears(&args.Key)
	for _, c := range kNears {
		*reply = append(*reply, *c)
	}
	return nil
}

func (rpc *RPCServer) FindValue(args *FindValueArgs, reply *FindValueResponse) error {
	server.Buckets.Update(&args.Contact)
	if err := server.Storage.Get(hex.EncodeToString(args.Key[:]), &reply.Value); err == skv.ErrNotFound {
		kNears := server.Buckets.KNears(&args.Key)
		for _, c := range kNears {
			reply.KNears = append(reply.KNears, *c)
		}
	}
	return nil
}

func (rpc *RPCServer) FindValueNetwork(args *FindValueArgs, reply *FindValueResponse) error {
	log.Printf("FIND_VALUE_NETWORK Key %s\n", hex.EncodeToString(args.Key[:]))
	nodes := server.LookUp(&args.Key)
	for _, node := range nodes {
		resp := SendFindValue(&server.Contact, node, &args.Key)
		if len(resp.Value) != 0 && len(resp.KNears) == 0 {
			*reply = resp
			break
		}
	}
	return nil
}

func SendPing(from *Contact, to *Contact) bool {
	args := PingArgs{Contact: *from}
	reply := false
	client, err := clientsManager.GetClient(to)
	if err == nil {
		err = client.Call("RPCServer.Ping", &args, &reply)
		if err != nil {
			reply = false
		}
	}
	return reply
}

func SendPingFromClient(to *Contact) bool {
	args := PingFromClientArgs{}
	reply := false
	client, err := clientsManager.GetClient(to)
	if err == nil {
		err = client.Call("RPCServer.Ping", &args, &reply)
		if err != nil {
			reply = false
		}
	}
	return reply
}

func SendStore(from *Contact, c *Contact, key Key, value []byte) bool {
	args := &StoreArgs{
		Contact: *from,
		Key:     key,
		Value:   value,
	}
	reply := false
	client, err := clientsManager.GetClient(c)
	if err == nil {
		err = client.Call("RPCServer.Store", &args, &reply)
		if err != nil {
			reply = false
		}
	}
	return reply
}

func SendFindNode(from *Contact, c *Contact, key *Key) []Contact {
	args := &FindNodeArgs{
		Contact: *from,
		Key:     *key,
	}
	var reply []Contact
	client, err := clientsManager.GetClient(c)
	if err == nil {
		client.Call("RPCServer.FindNode", &args, &reply)
	}
	return reply
}

func SendFindValue(from *Contact, c *Contact, key *Key) FindValueResponse {
	args := FindValueArgs{
		Contact: *from,
		Key:     *key,
	}
	var reply FindValueResponse
	client, err := clientsManager.GetClient(c)
	if err == nil {
		client.Call("RPCServer.FindValue", &args, &reply)
	}
	return reply
}

func SendStoreNetwork(c *Contact, key *Key, value []byte) bool {
	log.Printf("--> %s:%d STORE_NETWORK Key: %s\n", c.Ip.String(), c.Port, hex.EncodeToString(key[:]))
	args := &StoreArgs{
		Contact: Contact{},
		Key:     *key,
		Value:   value,
	}
	var reply bool
	client, err := clientsManager.GetClient(c)
	if err == nil {
		if err = client.Call("RPCServer.StoreNetwork", &args, &reply); err != nil {
			return false
		}
	}
	return reply
}

func SendFindValueNetwork(c *Contact, key *Key) []byte {
	log.Printf("--> %s:%d FIND_VALUE_NETWORK Key: %s\n", c.Ip.String(), c.Port, hex.EncodeToString(key[:]))
	args := &FindValueArgs{
		Contact: Contact{},
		Key:     *key,
	}
	var reply FindValueResponse
	client, err := clientsManager.GetClient(c)
	if err == nil {
		if err = client.Call("RPCServer.FindValueNetwork", &args, &reply); err != nil {
			return []byte{}
		}
	}
	return reply.Value
}

type LookUpContact struct {
	Contact *Contact
	Visit   bool
}

func InsertLC(set *[]*LookUpContact, c *LookUpContact) {
	left := 0
	rigth := len(*set) - 1
	for left != rigth {
		middle := (left + rigth) / 2
		if (*set)[middle].Contact.Id.Compare(&c.Contact.Id) == -1 {
			left = middle + 1
		} else {
			rigth = middle
		}
	}
	if (*set)[left].Contact.Id.Compare(&c.Contact.Id) != 0 {
		*set = append(*set, c)
	}
}

func (server Server) LookUp(key *Key) []*Contact {
	result := []*LookUpContact{
		{
			Contact: &server.Contact,
			Visit:   false,
		},
	}

	for {
		tmp := 3
		channels := make([]chan []Contact, 3)
		for i := 0; i < len(channels); i++ {
			channels[i] = make(chan []Contact)
		}
		for i := 0; tmp != 0 && i < len(result); i++ {
			if !result[i].Visit {
				go func(t, k int) {
					channels[t-1] <- SendFindNode(&server.Contact, result[k].Contact, key)
				}(tmp, i)
				tmp -= 1
				result[i].Visit = true
			}
		}
		if tmp == 3 {
			r := make([]*Contact, len(result))
			for i, v := range result {
				r[i] = v.Contact
			}
			return r
		}
		for tmp != 3 {
			select {
			case c := <-channels[0]:
				for _, v := range c {
					InsertLC(&result, &LookUpContact{
						Contact: &v,
						Visit:   false,
					})
				}
				tmp += 1
			case c := <-channels[1]:
				for _, v := range c {
					InsertLC(&result, &LookUpContact{
						Contact: &v,
						Visit:   false,
					})
				}
				tmp += 1
			case c := <-channels[2]:
				for _, v := range c {
					InsertLC(&result, &LookUpContact{
						Contact: &v,
						Visit:   false,
					})
				}
				tmp += 1
			}
		}

		sort.Slice(result, func(i, j int) bool {
			distI := result[i].Contact.Id.DistanceTo(key)
			distJ := result[j].Contact.Id.DistanceTo(key)
			return distI.Compare(&distJ) == -1
		})
		if len(result) > KSIZE {
			result = result[:KSIZE]
		}
	}
}
