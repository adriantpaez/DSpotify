package main

import (
	http_server "DSpotify/src/http-server"
	"DSpotify/src/kademlia"
	"crypto/sha1"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

func main() {
	var idSeed, database, ipArg string
	var inPort, outPort, httpPort int
	var err error
	for i := 1; i < len(os.Args)-1; {
		switch os.Args[i] {
		case "--idSeed":
			idSeed = os.Args[i+1]
		case "--inPort":
			inPort, err = strconv.Atoi(os.Args[i+1])
			if err != nil {
				log.Println(err.Error())
				return
			}
		case "--outPort":
			outPort, err = strconv.Atoi(os.Args[i+1])
			if err != nil {
				log.Println(err.Error())
				return
			}
		case "--httpPort":
			httpPort, err = strconv.Atoi(os.Args[i+1])
			if err != nil {
				log.Println(err.Error())
				return
			}
		case "--database":
			database = os.Args[i+1]
		case "--ip":
			ipArg = os.Args[i+1]
		default:
			log.Printf("Unexpected param %s\n", os.Args[i])
			return
		}
		i = i + 2
	}
	if idSeed == "" {
		println("Param --idSeed is required")
		return
	}
	if inPort == 0 {
		inPort = 8000
	}
	if outPort == 0 {
		outPort = 8001
	}
	if httpPort == 0 {
		httpPort = 8080
	}
	if database == "" {
		database = "database.db"
	}
	if ipArg == "" {
		ipArg = "127.0.0.1"
	}
	ip := net.ParseIP(ipArg)
	if ip == nil {
		fmt.Printf("ERROR: Invalid IP: %s\n", ipArg)
		return
	}
	//runtime.GOMAXPROCS(9)
	var key kademlia.Key = sha1.Sum([]byte(idSeed))
	server := kademlia.NewServer(key, ip, inPort, outPort, database)
	httpServer := http_server.HttpServer{
		Server: *server,
		Host:   ip,
		Port:   httpPort,
	}
	go httpServer.Start()
	server.Start()

}
