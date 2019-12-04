package main

import (
	httpserver "DSpotify/src/client/http-server"
	"context"
	"flag"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net"
)

type Client struct {
	Database *mongo.Database
}

func main() {
	ipArg := flag.String("ip", "127.0.0.1", "The IP of the client.")
	httpPort := flag.Int("port", 8081, "The port of the client")
	ip := net.ParseIP(*ipArg)
	clientOptions := options.Client().ApplyURI("mongodb://10.42.0.1:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	database := client.Database("dspotify")
	fmt.Printf("Hello client\n")
	httpServer := httpserver.HttpServer{
		Host:     ip,
		Port:     *httpPort,
		Database: database,
	}
	httpServer.Start()
}
