package httpserver

import (
	"DSpotify/src/client/db"
	"DSpotify/src/kademlia"
	"crypto/sha1"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"io"
	"log"
	"time"
)

var stalker NetworkStalker

type NetworkStalker struct {
	CurrentBlocks chan *ReaderCloserBlock
	Lock          bool
}

type ReaderCloserBlock struct {
	Index int
	Data  []byte
}

func (rc *ReaderCloserBlock) Read(p []byte) (int, error) {
	if rc.Index < len(rc.Data) {
		n := copy(p, rc.Data[rc.Index:])
		rc.Index += n
		return n, nil
	} else {
		return 0, io.EOF
	}
}

func (rc *ReaderCloserBlock) Close() error {
	fmt.Println("CLOSE")
	return nil
}

func (ns *NetworkStalker) Stalk(metadata db.SongArtistName, errChan *chan error) {
	println(ns)
	for i := 0; i < metadata.Song.Blocks; i++ {
		var key kademlia.Key = sha1.Sum([]byte(fmt.Sprintf("%s-%s-%d", metadata.Song.Title, metadata.Artist, i)))
		value, err := kademlia.SendFindValueNetwork(server.Kademlia, &key)
		*errChan <- err
		if len(value) == 0 {
			fmt.Println("VALUE NOT FOUND")
		}
		ns.CurrentBlocks <- &ReaderCloserBlock{
			Index: 0,
			Data:  value,
		}
	}
}

func (ns *NetworkStalker) Next() *ReaderCloserBlock {
	result := <-ns.CurrentBlocks
	return result
}

func (ns *NetworkStalker) play() {
	ns.Lock = false
}

func (ns *NetworkStalker) pause() {
	ns.Lock = true
}

func (q *Queue) DecodeAndFill() {
	streamer, format, err := mp3.Decode(stalker.Next())
	if err != nil {
		fmt.Println(err)
	}
	resampled := beep.Resample(4, format.SampleRate, 44100, streamer)
	q.Add(resampled)
}

type Queue struct {
	streamers []beep.Streamer
}

func (q *Queue) Add(streamers ...beep.Streamer) {
	q.streamers = append(q.streamers, streamers...)
}

func Play() {
	stalker.play()
}
func Pause() {
	stalker.pause()
}

func StreamInit(metadata db.SongArtistName) error {
	stalker = NetworkStalker{
		CurrentBlocks: make(chan *ReaderCloserBlock, 3),
		Lock:          false,
	}
	var errChan chan error
	go stalker.Stalk(metadata, &errChan)
	if err := <-errChan; err != nil {
		return err
	}
	sr := beep.SampleRate(44100)
	err := speaker.Init(sr, sr.N(time.Second/10))
	if err != nil {
		log.Fatal(err)
	}
	var queue Queue
	queue.DecodeAndFill()
	speaker.Play(&queue)
	time.Sleep(3 * time.Minute)
	return nil
}

func (q *Queue) Stream(samples [][2]float64) (n int, ok bool) {
	filled := 0
	for filled < len(samples) {
		if len(q.streamers) == 0 || stalker.Lock {
			for i := range samples[filled:] {
				samples[i][0] = 0
				samples[i][1] = 0
			}
			break
		}
		n, ok := q.streamers[0].Stream(samples[filled:])
		if !ok {
			q.DecodeAndFill()
			q.streamers = q.streamers[1:]
		}
		filled += n
	}
	return len(samples), true
}

func (q *Queue) Err() error {
	return nil
}
