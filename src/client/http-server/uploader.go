package httpserver

import (
	"DSpotify/src/client/db"
	"DSpotify/src/kademlia"
	"crypto/sha1"
	"fmt"
	"github.com/hajimehoshi/go-mp3"
	"github.com/viert/go-lame"
	"net"
	"os"
	"strings"
	"syscall"
)

type Block []byte

const CONNECTIONREFUSED = "Connection refused"
const UNKNOWNHOST = "Unknown host"
const TIMEOUT = "Timeout"

func (block *Block) Write(p []byte) (int, error) {
	*block = append(*block, p...)
	return len(p), nil
}

func UploadInit(metadata *db.SongArtistName, filepath string) error {
	if filepath != "" {
		file, err := os.Open(filepath)
		if err != nil {
			fmt.Println("ERROR:", err.Error())
			return err
		} else {
			stats, err := file.Stat()
			if err != nil {
				return err
			}
			if !stats.IsDir() {
				if strings.Contains(file.Name(), ".mp3") {
					return upload(file, metadata)
				} else {
					fmt.Println("ERROR:", err.Error())
					return err
				}
			} else {
				names, err := file.Readdirnames(0)
				if err != nil {
					return err
				}
				for _, n := range names {
					if strings.Contains(n, ".mp3") {
						newName := strings.Join([]string{filepath, n}, "/")
						newFile, err := os.Open(newName)
						if err != nil {
							return err
						}
						return upload(newFile, metadata)
					}
				}
			}
		}
	}
	return nil
}

func checkErr(err error) string {
	if err == nil {
		return "OK"

	} else if netError, ok := err.(net.Error); ok && netError.Timeout() {
		return TIMEOUT
	}
	switch t := err.(type) {
	case *net.OpError:
		if t.Op == "dial" {
			return UNKNOWNHOST
		} else if t.Op == "read" {
			return CONNECTIONREFUSED
		}

	case syscall.Errno:
		if t == syscall.ECONNREFUSED {
			return CONNECTIONREFUSED
		}
	}
	return "UNKNOWN"
}

func (block Block) save(metadata *db.SongArtistName) error {
	var key kademlia.Key = sha1.Sum([]byte(fmt.Sprintf("%s-%s-%d", metadata.Song.Title, metadata.Artist, metadata.Song.Blocks)))
	err := kademlia.SendStoreNetwork(server.Kademlia, &key, block)
	if err != nil && strings.Contains(err.Error(), "connection refused") {
		return ServerError{ErrorMessage: CONNECTIONREFUSED}
	}
	return err
}

func upload(file *os.File, metadata *db.SongArtistName) error {
	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		return err
	}
	buffer := make([]byte, 5000)
	total := 3000000
	var b *Block
	var encoder *lame.Encoder
	for {
		if n, err := decoder.Read(buffer); n > 0 {
			if err != nil {
				return err
			}
			if total >= 3000000 {
				if b != nil {
					err = b.save(metadata)
					if err != nil {
						return err
					}
					metadata.Song.Blocks++
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
				return err
			}
			total += n
		} else {
			break
		}
	}
	return nil
}
