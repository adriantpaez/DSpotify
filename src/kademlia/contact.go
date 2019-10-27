package kademlia

import (
	"net"
)

type Contact struct {
	Id   Key
	Ip   net.IP
	Port int
}

func NewContact(id Key, ip net.IP, port int) *Contact {
	return &Contact{Id: id, Ip: ip, Port: port}
}

func (c *Contact) Compare(other *Contact) int {
	return c.Id.Compare(&other.Id)
}
