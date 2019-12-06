package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type Db struct {
	Ip   string
	Port int
}

var db *mongo.Database

func dbs(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	switch req.Method {
	case http.MethodGet:
		fmt.Println("GET")
		coll := db.Collection("dbs")
		cur, err := coll.Aggregate(context.TODO(), bson.D{{"$sample", 1}})
		if err != nil {
			resp, _ := json.Marshal(false)
			w.Write(resp)
			return
		}
		db := Db{}
		err = cur.Decode(&db)
		if err != nil {
			resp, _ := json.Marshal(false)
			w.Write(resp)
			return
		}
		resp, _ := json.Marshal(&db)
		w.Write(resp)
	case http.MethodPost:
		fmt.Println("POST")
		ip := req.Form.Get("ip")
		port := req.Form.Get("port")
		_ip := net.ParseIP(ip)
		_port, err := strconv.Atoi(port)
		if _ip == nil || err != nil {
			resp, _ := json.Marshal(false)
			w.Write(resp)
			return
		}
		coll := db.Collection("dbs")
		_, err = coll.InsertOne(context.TODO(), bson.D{{"ip", ip}, {"port", _port}})
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
	clientOptions := options.Client().ApplyURI(strings.Join([]string{"mongodb://", net.JoinHostPort(*ip, strconv.Itoa(*port))}, ""))
	client, err := mongo.Connect(context.TODO(), clientOptions)
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
		log.Fatal(err)
	}

	http.HandleFunc("/dbs", dbs)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", *ip, *port), nil); err != nil {
		log.Fatal(err)
	}
}
