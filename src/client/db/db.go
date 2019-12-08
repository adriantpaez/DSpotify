package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type Artist struct {
	Name string
}

type Song struct {
	Artist interface{}
	Title  string
	Blocks int
}

type SongArtistName struct {
	Artist string
	Song   Song
}

func StoreArtist(client *mongo.Database, artist Artist) (*mongo.InsertOneResult, error) {
	collection := getCollection(client, "artist")
	return collection.InsertOne(context.Background(), artist)
}

func FindSongByArtist(client *mongo.Database, ArtistId interface{}) ([]Song, error) {
	collection := getCollection(client, "song")
	cursor, err := collection.Find(context.Background(), bson.D{{"artistId", ArtistId}})
	if err != nil {
		return nil, err
	}
	var result []Song
	err = cursor.All(context.Background(), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func listCollection(client *mongo.Database, name string) ([]string, error) {
	collection := getCollection(client, name)
	fmt.Println(collection.Name())
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}
	var result []Artist
	err = cursor.All(context.TODO(), &result)
	if err != nil {
		log.Fatal(err)
	}
	var names []string
	for _, item := range result {
		names = append(names, item.Name)
		fmt.Println(item.Name)
	}
	return names, nil
}

func ListArtists(client *mongo.Database) ([]string, error) {
	return listCollection(client, "artist")
}

func StoreSong(client *mongo.Database, song Song) (*mongo.InsertOneResult, error) {
	collection := getCollection(client, "song")
	return collection.InsertOne(context.Background(), song)
}

func getCollection(client *mongo.Database, coll string) *mongo.Collection {
	collection := client.Collection(coll)
	return collection
}
