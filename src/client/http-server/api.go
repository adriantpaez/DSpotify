package httpserver

import (
	"DSpotify/src/client/db"
	"DSpotify/src/kademlia"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net"
	"net/http"
)

type EndPoint struct {
	Path     string
	Function func(http.ResponseWriter, *http.Request)
}

var endpoints = []EndPoint{
	{
		Path:     "/findByArtist",
		Function: FindSongByArtistHandler,
	},
	{
		Path:     "/download",
		Function: DownloadSongHandler,
	},
	{
		Path:     "/play",
		Function: PlaySongHandler,
	},
	{
		Path:     "/pause",
		Function: PauseSongHandler,
	},
	{
		Path:     "/playSong",
		Function: DownloadSongHandler,
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

func PlaySongHandler(w http.ResponseWriter, req *http.Request) {
	Play()
}
func PauseSongHandler(w http.ResponseWriter, req *http.Request) {
	Pause()
}
func DownloadSongHandler(w http.ResponseWriter, req *http.Request) {
	title := getParameterFromQuery(req, "song")
	artist := getParameterFromQuery(req, "artist")
	coll := server.Database.Collection("song")
	hexId, _ := primitive.ObjectIDFromHex(artist)
	result := coll.FindOne(context.TODO(), bson.D{{"title", title}, {"artist", hexId}})
	var song db.Song
	err := result.Decode(&song)
	if err != nil {
		sendResponse(nil, err, &w)
		return
	}
	var art db.Artist
	coll = server.Database.Collection("artist")
	result = coll.FindOne(context.TODO(), bson.D{{"_id", hexId}})
	err = result.Decode(&art)
	if err != nil {
		sendResponse(nil, err, &w)
		return
	}
	err = StreamInit(db.SongArtistName{
		Artist: art.Name,
		Song:   song,
	})
}

func StoreArtistHandler(w http.ResponseWriter, req *http.Request) {
	artist := getParameterFromQuery(req, "artist")
	result, err := db.StoreArtist(server.Database, db.Artist{Name: artist})
	sendResponse(result, err, &w)
}

func StoreSongHandler(w http.ResponseWriter, req *http.Request) {
	title := getParameterFromQuery(req, "song")
	filepath := getParameterFromQuery(req, "file")
	artistId := getParameterFromQuery(req, "artist")
	hexId, _ := primitive.ObjectIDFromHex(artistId)
	coll := server.Database.Collection("artist")
	var artist db.Artist
	artistResult := coll.FindOne(context.TODO(), bson.D{{"_id", hexId}})
	err := artistResult.Decode(&artist)
	if err != nil {
		log.Fatal("Not artist found")
	} else {
		fmt.Println(artist.Name)
	}
	song := db.SongArtistName{
		Artist: artist.Name,
		Song: db.Song{
			Artist: hexId,
			Title:  title,
			Blocks: 0,
		},
	}
	//cosa nueva
	err = UploadInit(&song, filepath)
	var result *mongo.InsertOneResult
	if err == nil {
		result, err = db.StoreSong(server.Database, song.Song)
		if err != nil {
			log.Fatal(err)
		}
	}
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
	result, err := db.ListArtists(server.Database)
	sendResponse(result, err, &w)
}

func FindSongByArtistHandler(w http.ResponseWriter, req *http.Request) {
	artistId := getParameterFromQuery(req, "artist")
	hexId, _ := primitive.ObjectIDFromHex(artistId)
	result, err := db.FindSongByArtist(server.Database, hexId)
	sendResponse(result, err, &w)
}

func getParameterFromQuery(req *http.Request, key string) string {
	value, ok := req.URL.Query()[key]
	if !ok || len(value[0]) < 0 {
		log.Fatalf("Parameter %s not in query", key)
	}
	return value[0]
}

func GetNodes(ip *net.IP, port int) []kademlia.Contact {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/nodes", ip.String(), port))
	if err != nil {
		return []kademlia.Contact{}
	}
	data := []byte{}
	buffer := make([]byte, 100)
	for {
		if n, _ := resp.Body.Read(buffer); n > 0 {
			data = append(data, buffer[:n]...)
		} else {
			break
		}
	}
	nodes := []kademlia.Contact{}
	err = json.Unmarshal(data, &nodes)
	if err == nil {
		return nodes
	}
	return []kademlia.Contact{}
}
