package main

import (
	httpserver "DSpotify/src/client/http-server"
	"DSpotify/src/kademlia"
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	ipArg := flag.String("ip", "127.0.0.1", "The IP of the client.")
	httpPort := flag.Int("port", 8081, "The port of the client")
	DatabaseIp := flag.String("dbIp", "", "IP of known mongo Database server")
	var knownFile = flag.String("known", "", "File with known contact for join to network.")
	kademliaOutPort := flag.Int("kdOutPort", 0, "Port from which to connect to Kadamlia")
	DatabasePort := flag.Int("dbPort", 0, "Port of known Database server")
	flag.Parse()
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
	database, errD := establishDatabaseConnection(DatabaseIp, DatabasePort)
	if errD != nil {
		log.Println(errD.Error())
	}
	knownContact, postman, errK := establishKademliaConnection(ip, kademliaOutPort, knownFile)
	if errK != nil {
		log.Println(errK.Error())
	}
	if errD == nil && errK == nil {
		log.Printf("HTTP-SERVER SUCESSFULLY CONNECTED TO METADATA DATABASE dspotify\n AND KADEMLIA SERVER %s AT %s:%d", hex.EncodeToString(knownContact.Id[:]), knownContact.Ip, knownContact.Port)
	}
	httpServer := httpserver.HttpServer{
		Host:     ip,
		Port:     *httpPort,
		Database: database,
		Kademlia: knownContact,
		Postman:  postman,
	}
	httpServer.Start()
}

func establishDatabaseConnection(DatabaseIp *string, DatabasePort *int) (*mongo.Database, error) {
	clientOptions := options.Client().ApplyURI(strings.Join([]string{"mongodb://", net.JoinHostPort(*DatabaseIp, strconv.Itoa(*DatabasePort))}, ""))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, httpserver.ServerError{"HTTP-SERVER COULD NOT CONNECT TO DATABASE SERVER"}
	}
	return client.Database("dspotify"), nil
}

func establishKademliaConnection(clientIp net.IP, kademliaOutPort *int, knownFile *string) (*kademlia.Contact, *kademlia.Postman, error) {
	var knownContact *kademlia.Contact
	if *knownFile != "" {
		file, err := os.Open(*knownFile)
		if err != nil {
			return nil, nil, httpserver.ServerError{ErrorMessage: err.Error()}
		}
		info, err := file.Stat()
		if err != nil {
			return nil, nil, httpserver.ServerError{ErrorMessage: err.Error()}
		}
		data := make([]byte, info.Size())
		_, err = file.Read(data)
		if err != nil {
			return nil, nil, httpserver.ServerError{ErrorMessage: err.Error()}
		}
		c := kademlia.Contact{}
		err = json.Unmarshal(data, &c)
		if err != nil {
			return nil, nil, httpserver.ServerError{ErrorMessage: err.Error()}
		}
		knownContact = &c
	}
	postman := kademlia.NewPostman(100, clientIp, *kademliaOutPort)
	go postman.Start()
	if !kademlia.SendPing(nil, knownContact, postman) {
		return knownContact, postman, httpserver.ServerError{ErrorMessage: "HTTP SERVER COULD NOT CONNECT TO KADEMLIA NODE"}
	}
	return knownContact, postman, nil
}
