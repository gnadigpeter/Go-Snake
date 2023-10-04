package snake

type Node struct {
	coordinates Coordinate
	gCost       int
	hCost       int
	fCost       int
	parent      *Node
	isWalkable  bool
	direction   int
	isClosed    bool
}

func CreateNode(coordinate Coordinate, gCost int, hCost int) *Node {
	var node Node
	node.coordinates = coordinate
	node.gCost = 0
	node.hCost = hCost
	node.direction = -1
	node.calcFCost()
	return &node
}

func (n *Node) calcFCost() {
	n.fCost = n.gCost + n.hCost
}

func (n *Node) GetDrirection() int {
	return n.direction
}

func (n *Node) SetDrirection(dir int) {
	n.direction = dir
}

func (n *Node) GetParent() *Node {
	return n.parent
}

func (n *Node) SetParent(parent *Node) {
	if n.direction == -1 {
		if parent.coordinates.x-n.coordinates.x > 0 {
			//Left
			n.direction = 1
		} else if parent.coordinates.x-n.coordinates.x < 0 {
			//Right
			n.direction = 2
		} else {
			if parent.coordinates.y-n.coordinates.y > 0 {
				//up
				n.direction = 0
			} else {
				//down
				n.direction = 3
			}
		}
	}
	n.parent = parent
}

func (n *Node) SetgCost(gCost int) {
	n.gCost = gCost
	n.calcFCost()
}

func (n *Node) GetFCost() int {
	return n.fCost
}

func (n *Node) GetgCost() int {
	return n.gCost
}

func (n *Node) IsClosedd() bool {
	return n.isClosed
}

func (n *Node) Close() {
	n.isClosed = true
}

func (n *Node) compareTo(o *Node) int {
	if n.fCost > o.fCost {
		return 1
	} else if n.fCost < o.fCost {
		return -1
	}
	return 0
}

func (n *Node) Same(node *Node) bool {
	if n.coordinates.x == node.coordinates.x && n.coordinates.y == node.coordinates.y {
		return true
	}
	return false
}

func (n *Node) equals(o *Node) bool {
	if n == o {
		return true
	}
	if o == nil {
		return false
	}
	return n.coordinates.x == o.coordinates.x && n.coordinates.y == o.coordinates.y
}
