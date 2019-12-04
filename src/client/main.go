package main

import (
	httpserver "DSpotify/src/client/http-server"
	"DSpotify/src/kademlia"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type Client struct {
	Database *mongo.Database
}

func main() {
	ipArg := flag.String("ip", "127.0.0.1", "The IP of the client.")
	httpPort := flag.Int("port", 8081, "The port of the client")
	DatabaseIp := flag.String("dbIp", "", "IP of known mongo Database server")
	var knownFile = flag.String("known", "", "File with known contact for join to network.")
	kademliaOutPort := flag.Int("kdOutPort", 0, "Port from which to connect to Kadamlia")
	DatabasePort := flag.Int("dbPort", 0, "Port of known Database server")
	if *DatabaseIp == "" {
		log.Fatal("Empty Database IP")
	}
	if *DatabasePort == 0 {
		log.Fatal("Empty Database port")
	}
	if *knownFile == "" {
		log.Fatal("Empty known ontact file")
	}
	if *kademliaOutPort == 0 {
		log.Fatal("Empty Kademlia out port")
	}
	ip := net.ParseIP(*ipArg)
	clientOptions := options.Client().ApplyURI(strings.Join([]string{"mongodb://", net.JoinHostPort(*DatabaseIp, strconv.Itoa(*DatabasePort))}, ""))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal("HTTP-SERVER COULD NOT CONNECT TO DATABASE SERVER")
	}
	var knownContact *kademlia.Contact
	if *knownFile != "" {
		file, err := os.Open(*knownFile)
		if err != nil {
			fmt.Println("ERROR:", err.Error())
			return
		}
		info, err := file.Stat()
		if err != nil {
			fmt.Println("ERROR:", err.Error())
			return
		}
		data := make([]byte, info.Size())
		_, err = file.Read(data)
		if err != nil {
			fmt.Println("ERROR:", err.Error())
			return
		}
		c := kademlia.Contact{}
		err = json.Unmarshal(data, &c)
		if err != nil {
			fmt.Println("ERROR:", err.Error())
			return
		}
		knownContact = &c
	}
	postman := kademlia.NewPostman(100, *kademliaOutPort)
	pingResponse := kademlia.SendPing(nil, knownContact, postman)
	if !pingResponse {
		log.Fatal("HTTP-SERVER COULD NOT CONNECT TO KADEMLIA NODE")
	}
	log.Printf("HTTP-SERVER SUCESSFULLY CONNECTED METADATA DATABASE dspotify\n AND KADEMLIA SERVER %s AT %s:%d", knownContact.Id, knownContact.Ip, knownContact.Port)

	database := client.Database("dspotify")
	httpServer := httpserver.HttpServer{
		Host:     ip,
		Port:     *httpPort,
		Database: database,
		Kademlia: knownContact,
	}
	httpServer.Start()
}
