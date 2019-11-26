package http_server

import (
	"DSpotify/src/kademlia"
	"log"
	"net"
	"net/http"
)

type HttpServer struct {
	Server kademlia.Server
	Host   net.IP
	Port   int
}

var kademliaServer kademlia.Server

func (server HttpServer) Start() {
	kademliaServer = server.Server
	for i := 0; i < len(endpoints); i++ {
		http.HandleFunc(endpoints[i].Path, endpoints[i].Function)
	}
	addr := net.TCPAddr{
		IP:   server.Host,
		Port: server.Port,
		Zone: "",
	}
	err := http.ListenAndServe(addr.String(), nil)
	if err != nil {
		log.Printf("HTTP-SERVER ERROR: %s\n", err.Error())
	}
}
