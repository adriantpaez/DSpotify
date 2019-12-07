package main

import (
	http_server "DSpotify/src/http-server"
	"DSpotify/src/kademlia"
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/skycoin/skycoin/src/api"
	"log"
	"net"
	"net/http"
	"os"
)

func registerNode(server *kademlia.Contact, ip *string, port *int) bool {
	data, err := json.Marshal(server)
	if err != nil {
		return false
	} else {
		_, err := http.Post(fmt.Sprintf("http://%s:%d/nodes", *ip, *port), api.ContentTypeJSON, bytes.NewBuffer(data))
		if err != nil {
			return false
		}
	}
	return true
}

func main() {
	var idSeed = flag.String("idSeed", "", "[REQUIRED] Seed for NodeID. NodeID=sha1(idSeed).")
	var database = flag.String("database", "database.db", "Path for database.")
	var ipArg = flag.String("ip", "127.0.0.1", "The IP of the node.")
	var knownFile = flag.String("known", "", "File with known contact for join to network.")
	var inPort = flag.Int("inPort", 8000, "Input port for listen all RPC requests.")
	var httpPort = flag.Int("httpPort", 8080, "Port for listen requests to web API.")
	var trackerIp = flag.String("trackerIp", "127.0.0.1", "The IP of the tracker")
	var trackerPort = flag.Int("trackerPort", 7000, "The Port of the tracker")
	flag.Parse()

	ip := net.ParseIP(*ipArg)
	if ip == nil {
		fmt.Printf("ERROR: Invalid IP: %s\n", *ipArg)
		return
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
		if _, err := file.Read(data); err == nil {
			c := kademlia.Contact{}
			err = json.Unmarshal(data, &c)
			if err == nil {
				knownContact = &c
			} else {
				fmt.Println("ERROR:", err.Error())
			}
		}
	}

	if *idSeed == "" {
		fmt.Println("Not idSeed set.")
		return
	}

	var key kademlia.Key = sha1.Sum([]byte(*idSeed))
	server := kademlia.NewServer(key, ip, *inPort, *database)
	httpServer := http_server.HttpServer{
		Server: *server,
		Host:   ip,
		Port:   *httpPort,
	}
	go httpServer.Start()
	if registerNode(&server.Contact, trackerIp, trackerPort) {
		log.Println("Register done")
	} else {
		log.Println("Register fail")
	}
	server.Start(knownContact)
}
