package kademlia

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

func (b *Bucket) Update(c *Contact) {
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
	} else if !SendPing(&server.Contact, c, server.Postman) {
		updated := (*b)[1:]
		updated = append(updated, c)
		*b = updated
	}
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

func (table BucketsTable) Update(c *Contact) {
	b := table.FindBucket(c)
	b.Update(c)
}

func insertBucket(root **AVLNode, key *Key, b *Bucket) int {
	for k := 0; k < len(*b); k++ {
		Insert(root, key.DistanceTo(&(*b)[k].Id), (*b)[k])
	}
	return len(*b)
}

func (table BucketsTable) KNears(key *Key) []*Contact {
	var root *AVLNode = nil
	count := 0
	dist := table.Owner.Id.DistanceTo(key)

	for i := 0; count <= KSIZE && i < KEYSIZE*8; i++ {
		b, _ := dist.GetBit(i)
		if b == 1 {
			count += insertBucket(&root, key, &table.Buckets[i])
		}
	}

	for i := KEYSIZE*8 - 1; count <= KSIZE && i >= 0; i-- {
		b, _ := dist.GetBit(i)
		if b == 0 {
			count += insertBucket(&root, key, &table.Buckets[i])
		}
	}

	for count > KSIZE {
		max := Max(root)
		Remove(&root, max.Key)
		count -= 1
	}
	resp := []*Contact{}
	iter := make(chan *Contact)
	go root.PreOrden(iter)
	for contact := range iter {
		resp = append(resp, contact)
	}
	return resp
}
