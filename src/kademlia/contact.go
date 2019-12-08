package kademlia

import (
	"encoding/hex"
	"fmt"
	"net"
)

type Contact struct {
	Id   Key
	Ip   net.IP
	Port int
}

func (c *Contact) String() string {
	return fmt.Sprintf("%s %s:%d", hex.EncodeToString(c.Id[:]), c.Ip.String(), c.Port)
}

func NewContact(id Key, ip net.IP, port int) *Contact {
	return &Contact{Id: id, Ip: ip, Port: port}
}

func (c *Contact) Compare(other *Contact) int {
	return c.Id.Compare(&other.Id)
}
