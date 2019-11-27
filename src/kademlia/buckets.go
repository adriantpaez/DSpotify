package kademlia

import (
	"encoding/hex"
	"fmt"
)

var server *Server

type Bucket []*Contact

func (b Bucket) Contains(c *Contact) int {
	for i := 0; i < len(b); i++ {
		if b[i].Compare(c) == 0 {
			return i
		}
	}
	return -1
}

func (b *Bucket) Update(c *Contact) bool {
	if i := b.Contains(c); i != -1 {
		updated := Bucket{}
		updated = append(updated, (*b)[i])
		for j := 0; j < len(*b); j++ {
			if j != i {
				updated = append(updated, (*b)[j])
			}
		}
		*b = updated
	} else if len(*b) < KSIZE {
		*b = append(*b, c)
	} else {
		// TODO: Run PING for first contact on b
		// If b[0] is alive
		// then -> forget c
		// else -> remove b[0] and append c to end of b
		return false
	}
	return true
}

type BucketsTable struct {
	Owner   Contact
	Buckets []Bucket
}

func NewBucketsTable(owner *Server) *BucketsTable {
	server = owner
	table := BucketsTable{
		Owner:   owner.Contact,
		Buckets: make([]Bucket, KEYSIZE*8),
	}
	for i := 0; i < KEYSIZE*8; i++ {
		table.Buckets[i] = Bucket{}
	}
	return &table
}

func (table BucketsTable) FindBucket(c *Contact) *Bucket {
	i := table.Owner.Id.CommonPrefixLength(&c.Id)
	return &table.Buckets[i]
}

func (table BucketsTable) Contains(c *Contact) bool {
	bucket := table.FindBucket(c)
	return bucket.Contains(c) != -1
}

func (table BucketsTable) Update(c *Contact) bool {
	b := table.FindBucket(c)
	return b.Update(c)
}

func (table BucketsTable) Print() {
	fmt.Println("BUCKETS TABLE")
	for i := 0; i < len(table.Buckets); i++ {
		bucket := table.Buckets[i]
		fmt.Printf("BUCKET %d\n", i)
		for j := 0; j < len(bucket); i++ {
			if bucket[i] == nil {
				break
			}
			fmt.Printf("%s  %s  %d\n", hex.EncodeToString(bucket[i].Id[:]), bucket[i].Ip.String(), bucket[i].Port)
		}
	}
}
