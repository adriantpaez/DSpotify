package kademlia

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
		for j := 0; j < len(*b); j++ {
			if j != i {
				updated = append(updated, (*b)[j])
			}
		}
		updated = append(updated, (*b)[i])
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

func NewBucketsTable(owner Contact) *BucketsTable {
	table := BucketsTable{
		Owner:   owner,
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
