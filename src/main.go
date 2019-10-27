package main

import (
	"DSpotify/src/kademlia"
	"encoding/hex"
	"log"
)

func main() {
	key := kademlia.Key{}
	key[11] = byte(123)
	server := kademlia.NewServer(key, []byte{127, 0, 0, 1}, 8000)
	log.Printf("%s %s:%d\n", hex.EncodeToString(key[:]), server.Self.Ip.String(), server.Self.Port)
	server.Start()
}
