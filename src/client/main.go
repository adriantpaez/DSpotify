package main

import (
	httpserver "DSpotify/src/client/http-server"
	"DSpotify/src/kademlia"
	"context"
	"encoding/hex"
	"flag"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
)

func main() {
	ipArg := flag.String("ip", "127.0.0.1", "The IP of the client.")
	httpPort := flag.Int("port", 8081, "The port of the client")
	DatabaseIp := flag.String("dbIp", "", "IP of known mongo Database server")
	networkTracker := flag.String("trackerIp", "", "IP of network tracker")
	DatabasePort := flag.Int("dbPort", 0, "Port of known Database server")
	flag.Parse()
	if *DatabaseIp == "" {
		log.Fatal("Empty Database IP")
	}
	if *DatabasePort == 0 {
		log.Fatal("Empty Database port")
	}
	if *networkTracker == "" {
		log.Fatal("Empty Tracker IP")
	}
	ip := net.ParseIP(*ipArg)
	networkTrackerIp := net.ParseIP(*networkTracker)
	database, errD := establishDatabaseConnection(DatabaseIp, DatabasePort)
	if errD != nil {
		log.Println(errD.Error())
	}
	knownContact, errK := establishKademliaConnection(&networkTrackerIp)
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
		Tracker:  &networkTrackerIp,
	}
	httpServer.Start()
}

func establishDatabaseConnection(DatabaseIp *string, DatabasePort *int) (*mongo.Database, error) {
	clientOptions := options.Client().ApplyURI(strings.Join([]string{"mongodb://", net.JoinHostPort(*DatabaseIp, strconv.Itoa(*DatabasePort))}, ""))
	client, _ := mongo.Connect(context.TODO(), clientOptions)
	err := client.Ping(context.TODO(), nil)
	if err != nil {
		log.Println("error")
		return nil, httpserver.ServerError{"HTTP-SERVER COULD NOT CONNECT TO DATABASE SERVER"}
	}
	log.Println("No error")
	return client.Database("dspotify"), nil
}

func establishKademliaConnection(ip *net.IP) (*kademlia.Contact, error) {
	var knownContact *kademlia.Contact
	for {
		nodes := httpserver.GetNodes(ip, 7000)
		if len(nodes) == 0 {
			log.Println("WARNING: Not nodes for join to network.")
			break
		} else {
			knownContact = &nodes[rand.Intn(len(nodes))]
		}
		kademlia.InitClientsManager()
		if kademlia.SendPingFromClient(knownContact) {
			break
		}
	}
	return knownContact, nil
}
