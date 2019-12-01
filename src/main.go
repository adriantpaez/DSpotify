package main

import (
	http_server "DSpotify/src/http-server"
	"DSpotify/src/kademlia"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
)

func main() {
	var idSeed = flag.String("idSeed", "", "[REQUIRED] Seed for NodeID. NodeID=sha1(idSeed).")
	var database = flag.String("database", "database.db", "Path for database.")
	var ipArg = flag.String("ip", "127.0.0.1", "The IP of the node.")
	var knownFile = flag.String("known", "", "File with known contact for join to network.")
	var inPort = flag.Int("inPort", 8000, "Input port for listen all RPC requests.")
	var outPort = flag.Int("outPort", 8001, "Output port for send all RPC requests.")
	var httpPort = flag.Int("httpPort", 8080, "Port for listen requests to web API.")
	flag.Parse()

	ip := net.ParseIP(*ipArg)
	if ip == nil {
		fmt.Printf("ERROR: Invalid IP: %s\n", ipArg)
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
		file.Read(data)
		c := kademlia.Contact{}
		err = json.Unmarshal(data, &c)
		if err != nil {
			fmt.Println("ERROR:", err.Error())
			return
		}
		knownContact = &c
	}

	if *idSeed == "" {
		fmt.Println("Not idSeed set.")
		return
	}

	var key kademlia.Key = sha1.Sum([]byte(*idSeed))
	server := kademlia.NewServer(key, ip, *inPort, *outPort, *database)
	httpServer := http_server.HttpServer{
		Server: *server,
		Host:   ip,
		Port:   *httpPort,
	}
	go httpServer.Start()
	server.Start(knownContact)
}
