package httpserver

import (
	"DSpotify/src/kademlia"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net"
	"net/http"
)

type HttpServer struct {
	Host     net.IP
	Port     int
	Database *mongo.Database
	Kademlia *kademlia.Contact
}

var server *HttpServer

func (thisServer HttpServer) Start() {
	server = &thisServer
	for i := 0; i < len(endpoints); i++ {
		http.HandleFunc(endpoints[i].Path, endpoints[i].Function)
	}

	fs := http.FileServer(http.Dir("interface"))
	http.Handle("/", fs)

	addr := net.TCPAddr{
		IP:   thisServer.Host,
		Port: thisServer.Port,
		Zone: "",
	}
	err := http.ListenAndServe(addr.String(), nil)
	if err != nil {
		log.Printf(ServerError{ErrorMessage: fmt.Sprintf("HTTP-SERVER ERROR: %s\n", err.Error())}.Error())
	}
}

type ServerError struct {
	ErrorMessage string
}

func (err ServerError) Error() string {
	return err.ErrorMessage
}
