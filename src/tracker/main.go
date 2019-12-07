package main

import (
	"DSpotify/src/kademlia"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/skycoin/skycoin/src/api"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var db *mongo.Database

func parseBodyDatabase(req *http.Request) net.TCPAddr {
	resp := net.TCPAddr{}
	data := []byte{}
	buffer := make([]byte, 100)
	for {
		if n, _ := req.Body.Read(buffer); n > 0 {
			data = append(data, buffer[:n]...)
		} else {
			break
		}
	}
	err := json.Unmarshal(data, &resp)
	if err != nil {
		log.Println(err.Error())
	}
	return resp
}

func parseBodyContact(req *http.Request) kademlia.Contact {
	resp := kademlia.Contact{}
	data := []byte{}
	buffer := make([]byte, 100)
	for {
		if n, _ := req.Body.Read(buffer); n > 0 {
			data = append(data, buffer[:n]...)
		} else {
			break
		}
	}
	err := json.Unmarshal(data, &resp)
	if err != nil {
		log.Println(err.Error())
	}
	return resp
}

func dbs(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", api.ContentTypeJSON)
	switch req.Method {
	case http.MethodGet:
		fmt.Println("GET /dbs")
		coll := db.Collection("dbs")
		cur, err := coll.Find(context.TODO(), bson.D{})
		if err != nil {
			log.Println(err.Error())
			resp, _ := json.Marshal(false)
			w.Write(resp)
			return
		}
		db := []net.TCPAddr{}
		err = cur.All(context.TODO(), &db)
		if err != nil {
			resp, _ := json.Marshal(false)
			w.Write(resp)
			return
		}
		resp, _ := json.Marshal(&db)
		w.Write(resp)
	case http.MethodPost:
		fmt.Println("POST /dbs")
		body := parseBodyDatabase(req)
		coll := db.Collection("dbs")
		_, err := coll.InsertOne(context.TODO(), body)
		if err != nil {
			resp, _ := json.Marshal(false)
			w.Write(resp)
		} else {
			resp, _ := json.Marshal(true)
			w.Write(resp)
		}
	}
}

func nodes(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", api.ContentTypeJSON)
	switch req.Method {
	case http.MethodGet:
		fmt.Println("GET /nodes")
		coll := db.Collection("nodes")
		cur, err := coll.Find(context.TODO(), bson.D{})
		if err != nil {
			log.Println(err.Error())
			resp, _ := json.Marshal(false)
			w.Write(resp)
			return
		}
		nodes := []kademlia.Contact{}
		err = cur.All(context.TODO(), &nodes)
		if err != nil {
			resp, _ := json.Marshal(false)
			w.Write(resp)
			return
		}
		resp, _ := json.Marshal(&nodes)
		w.Write(resp)
	case http.MethodPost:
		fmt.Println("POST /nodes")
		body := parseBodyContact(req)
		coll := db.Collection("nodes")
		_, err := coll.InsertOne(context.TODO(), body)
		if err != nil {
			resp, _ := json.Marshal(false)
			w.Write(resp)
		} else {
			resp, _ := json.Marshal(true)
			w.Write(resp)
		}
	}
}

func connectToDatabase(ip *string, port *int) (*mongo.Database, error) {
	fmt.Printf("Conneting to database on %s:%d - ", *ip, *port)
	clientOptions := options.Client().ApplyURI(strings.Join([]string{"mongodb://", net.JoinHostPort(*ip, strconv.Itoa(*port))}, ""))
	clientOptions.SetConnectTimeout(10 * time.Second)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	db := client.Database("tracker")
	return db, nil
}

func main() {
	var dbIp = flag.String("dbIp", "127.0.0.1", "IP of database")
	var dbPort = flag.Int("dbPort", 27017, "Port of database")
	var ip = flag.String("ip", "127.0.0.1", "IP of tracker server")
	var port = flag.Int("port", 7000, "Port of tracker server")
	flag.Parse()

	var err error

	if db, err = connectToDatabase(dbIp, dbPort); err != nil {
		fmt.Printf("fail\n")
		log.Fatal(err)
	}
	fmt.Printf("done\n")

	http.HandleFunc("/dbs", dbs)
	http.HandleFunc("/nodes", nodes)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", *ip, *port), nil); err != nil {
		log.Fatal(err)
	}
}
