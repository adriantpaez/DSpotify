package httpserver

import (
	"DSpotify/src/client/db"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"strconv"
)

type EndPoint struct {
	Path     string
	Function func(http.ResponseWriter, *http.Request)
}

var database *mongo.Database

var endpoints = []EndPoint{
	{
		Path:     "/findByArtist",
		Function: FindSongByArtistHandler,
	},
	{
		Path:     "/listArtists",
		Function: ListArtistsHandler,
	},
	{
		Path:     "/storeArtist",
		Function: StoreArtistHandler,
	},
	{
		Path:     "/storeSong",
		Function: StoreSongHandler,
	},
}

func StoreArtistHandler(w http.ResponseWriter, req *http.Request) {
	artist := getParameterFromQuery(req, "artist")
	result, err := db.StoreArtist(database, db.Artist{Name: artist})
	sendResponse(result, err, &w)
}

func StoreSongHandler(w http.ResponseWriter, req *http.Request) {
	title := getParameterFromQuery(req, "song")
	artist := getParameterFromQuery(req, "artist")
	blocks := getParameterFromQuery(req, "blocks")
	hexId, _ := primitive.ObjectIDFromHex(artist)
	intBlocks, _ := strconv.Atoi(blocks)
	result, err := db.StoreSong(database, db.Song{Title: title, KademliaBlocks: intBlocks, ArtistId: hexId})
	sendResponse(result, err, &w)

}

func sendResponse(result interface{}, err error, w *http.ResponseWriter) {
	var jsonResult []byte
	if err != nil {
		jsonResult, _ = json.Marshal(false)
	} else {
		jsonResult, err = json.Marshal(result)
	}
	(*w).Header().Add("Content-Type", "application/json")
	_, _ = (*w).Write(jsonResult)
}

func ListArtistsHandler(w http.ResponseWriter, req *http.Request) {
	result, err := db.ListArtists(database)
	sendResponse(result, err, &w)
}

func FindSongByArtistHandler(w http.ResponseWriter, req *http.Request) {
	artistId := getParameterFromQuery(req, "artistId")
	hexId, _ := primitive.ObjectIDFromHex(artistId)
	result, err := db.FindSongByArtist(database, hexId)
	sendResponse(result, err, &w)
}

func getParameterFromQuery(req *http.Request, key string) string {
	value, ok := req.URL.Query()[key]
	if !ok || len(value[0]) < 0 {
		log.Fatalf("Parameter %s not in query", key)
	}
	return value[0]
}
