package kademlia

type AVLNode struct {
	Key     Key
	Data    *Contact
	Balance int
	Link    [2]*AVLNode
}

func single(root *AVLNode, dir int) *AVLNode {
	save := root.Link[1-dir]
	root.Link[1-dir] = save.Link[dir]
	save.Link[dir] = root
	return save
}

func double(root *AVLNode, dir int) *AVLNode {
	save := root.Link[1-dir].Link[dir]
	root.Link[1-dir].Link[dir] = save.Link[1-dir]
	save.Link[1-dir] = root.Link[1-dir]
	root.Link[1-dir] = save
	save = root.Link[1-dir]
	root.Link[1-dir] = save.Link[dir]
	save.Link[dir] = root
	return save
}

func adjustBalance(root *AVLNode, dir, balance int) {
	n := root.Link[dir]
	nn := n.Link[1-dir]
	switch nn.Balance {
	case 0:
		root.Balance = 0
		n.Balance = 0
	case balance:
		root.Balance = -balance
		n.Balance = 0
	default:
		root.Balance = 0
		n.Balance = balance
	}
	nn.Balance = 0
}

func insertBalance(root *AVLNode, dir int) *AVLNode {
	n := root.Link[dir]
	balance := 2*dir - 1
	if n.Balance == balance {
		root.Balance = 0
		n.Balance = 0
		return single(root, 1-dir)
	}
	adjustBalance(root, dir, balance)
	return double(root, 1-dir)
}

func insertR(root *AVLNode, key Key, data *Contact) (*AVLNode, bool) {
	if root == nil {
		return &AVLNode{Key: key, Data: data}, false
	}
	dir := 0
	if root.Key.Compare(&key) == -1 {
		dir = 1
	}
	var done bool
	root.Link[dir], done = insertR(root.Link[dir], key, data)
	if done {
		return root, true
	}
	root.Balance += 2*dir - 1
	switch root.Balance {
	case 0:
		return root, true
	case 1, -1:
		return root, false
	}
	return insertBalance(root, dir), true
}

func Insert(tree **AVLNode, key Key, data *Contact) {
	*tree, _ = insertR(*tree, key, data)
}

func Remove(tree **AVLNode, key Key) {
	*tree, _ = removeR(*tree, key)
}

func removeR(root *AVLNode, key Key) (*AVLNode, bool) {
	if root == nil {
		return nil, false
	}
	if root.Key.Compare(&key) == 0 {
		switch {
		case root.Link[0] == nil:
			return root.Link[1], false
		case root.Link[1] == nil:
			return root.Link[0], false
		}
		heir := root.Link[0]
		for heir.Link[1] != nil {
			heir = heir.Link[1]
		}
		root.Key = heir.Key
		root.Data = heir.Data
		key = heir.Key
	}
	dir := 0
	if root.Key.Compare(&key) == -1 {
		dir = 1
	}
	var done bool
	root.Link[dir], done = removeR(root.Link[dir], key)
	if done {
		return root, true
	}
	root.Balance += 1 - 2*dir
	switch root.Balance {
	case 1, -1:
		return root, true
	case 0:
		return root, false

	}
	return removeBalance(root, dir)
}

func removeBalance(root *AVLNode, dir int) (*AVLNode, bool) {
	n := root.Link[1-dir]
	balance := 2*dir - 1
	switch n.Balance {
	case -balance:
		root.Balance = 0
		n.Balance = 0
		return single(root, dir), false
	case balance:
		adjustBalance(root, 1-dir, -balance)
		return double(root, dir), false
	}
	root.Balance = -balance
	n.Balance = balance
	return single(root, dir), true
}

func Max(root *AVLNode) *AVLNode {
	if root.Link[1] == nil {
		return root
	}
	return Max(root.Link[1])
}

func (root *AVLNode) PreOrden() <-chan *Contact {
	var ch = make(chan *Contact)
	go func() {
		for node := range root.Link[0].PreOrden() {
			ch <- node
		}
		ch <- root.Data
		for node := range root.Link[1].PreOrden() {
			ch <- node
		}
	}()
	return ch
}
