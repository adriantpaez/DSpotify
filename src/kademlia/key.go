package kademlia

import (
	"bytes"
	"fmt"
)

type Key [KEYSIZE]byte

func (key *Key) CommonPrefixLength(other *Key) int {
	for i := 0; i < KEYSIZE; i++ {
		if key[i] != other[i] {
			b := byte(1) << 7
			for j := uint(0); j <= 7; j++ {
				if key[i]&b != other[i]&b {
					return 8*i + int(j)
				}
				b = b >> 1
			}
		}
	}
	return len(key)*8 - 1
}

func (key *Key) DistanceTo(other *Key) Key {
	distance := Key{}
	for i := 0; i < KEYSIZE; i++ {
		distance[i] = key[i] ^ other[i]
	}
	return distance
}

func (key *Key) Compare(other *Key) int {
	return bytes.Compare(key[:], other[:])
}

func (key *Key) GetBit(index int) (int, error) {
	bytePos := index / 8
	offset := uint8(index % 8)
	if bytePos >= len(key) {
		return -1, fmt.Errorf("Bad index %d\n", index)
	}
	temp := byte(1) << (7 - offset)
	if key[bytePos]&temp == 0 {
		return 0, nil
	}
	return 1, nil
}
