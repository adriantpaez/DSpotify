package main

import (
	http_server "DSpotify/src/http-server"
	"DSpotify/src/kademlia"
	"crypto/sha1"
	"flag"
	"fmt"
	"net"
)

func main() {
	var idSeed = flag.String("idSeed", "", "[REQUIRED] Seed for NodeID. NodeID=sha1(idSeed).")
	var database = flag.String("database", "database.db", "Path for database.")
	var ipArg = flag.String("ip", "127.0.0.1", "The IP of the node.")
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
	tip := net.ParseIP(*trackerIp)
	server.Start(&tip, *trackerPort)
}
