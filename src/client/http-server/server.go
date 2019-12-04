package httpserver

import (
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net"
	"net/http"
)

type HttpServer struct {
	Host     net.IP
	Port     int
	Database *mongo.Database
}

func (server HttpServer) Start() {
	database = server.Database
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
