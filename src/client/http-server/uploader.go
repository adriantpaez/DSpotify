package httpserver

import (
	"DSpotify/src/client/db"
	"DSpotify/src/kademlia"
	"crypto/sha1"
	"fmt"
	"github.com/hajimehoshi/go-mp3"
	"github.com/viert/go-lame"
	"log"
	"os"
	"strings"
)

type Block []byte

func (block *Block) Write(p []byte) (int, error) {
	*block = append(*block, p...)
	return len(p), nil
}

func UploadInit(metadata *db.SongArtistName, filepath string) error {
	if filepath != "" {
		file, err := os.Open(filepath)
		if err != nil {
			fmt.Println("ERROR:", err.Error())
		} else {
			stats, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}
			if !stats.IsDir() {
				if strings.Contains(file.Name(), ".mp3") {
					upload(file, metadata)
				} else {
					log.Fatal("FILE MUST BE MP3")
				}
			} else {
				names, err := file.Readdirnames(0)
				if err != nil {
					log.Fatal(err)
				}
				for _, n := range names {
					if strings.Contains(n, ".mp3") {
						newName := strings.Join([]string{filepath, n}, "/")
						newFile, err := os.Open(newName)
						if err != nil {
							log.Fatal(err)
						}
						upload(newFile, metadata)
					}
				}
			}
		}
	}
	return nil
}

func (block Block) save(metadata *db.SongArtistName) error {
	var key kademlia.Key = sha1.Sum([]byte(fmt.Sprintf("%s-%s-%d", metadata.Song.Title, metadata.Artist, metadata.Song.Blocks)))
	if !kademlia.SendStoreNetwork(server.Kademlia, &key, block) {
		return ServerError{ErrorMessage: "STORE FAILED"}
	}
	return nil
}

func upload(file *os.File, metadata *db.SongArtistName) {
	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		log.Fatal(err)
	}
	buffer := make([]byte, 5000)
	total := 3000000
	var b *Block
	var encoder *lame.Encoder
	for {
		if n, err := decoder.Read(buffer); n > 0 {
			if err != nil {
				log.Fatal(err)
			}
			if total >= 3000000 {
				if b != nil {
					err = b.save(metadata)
					metadata.Song.Blocks++
					if err != nil {
						log.Fatal(err)
					}
				}
				if encoder != nil {
					encoder.Close()
				}
				b = &Block{}
				encoder = lame.NewEncoder(b)
				total = 0
			}
			n, err = encoder.Write(buffer[:n])
			if err != nil {
				log.Fatal(err)
			}
			total += n
		} else {
			break
		}
	}
}
