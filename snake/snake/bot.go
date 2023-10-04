package snake

// type Tree struct {
// 	root *Node
// }
// type Node struct {
// 	key   int
// 	left  *Node
// 	right *Node
// 	up    *Node
// 	down  *Node
// }

// func (t *Tree) insert(data int) {
// 	if t.root == nil {
// 		t.root = &Node{key: data}
// 	} else {
// 		t.root.insert(data)
// 	}
// }

// func (n *Node) insert(data int) {
// 	switch data {
// 	case 0:
// 		if n.up == nil {
// 			n.up = &Node{key: data}
// 		} else {
// 			n.up.insert(data)
// 		}
// 	case 1:
// 		if n.left == nil {
// 			n.left = &Node{key: data}
// 		} else {
// 			n.left.insert(data)
// 		}
// 	case 2:
// 		if n.right == nil {
// 			n.right = &Node{key: data}
// 		} else {
// 			n.right.insert(data)
// 		}
// 	case 3:
// 		if n.down == nil {
// 			n.down = &Node{key: data}
// 		} else {
// 			n.down.insert(data)
// 		}
// 	}

// }

// func getValues(n *Node, treeDir []int) []int {
// 	if n == nil {
// 		return nil
// 	} else {
// 		treeDir = append(treeDir, n.key)
// 		getValues(n.up, treeDir)
// 		getValues(n.left, treeDir)
// 		getValues(n.right, treeDir)
// 		getValues(n.down, treeDir)
// 	}
// 	return treeDir
// }

// // func getCurrentActiveNode(t *Node) int{
// // 	if t == nil{
// // 		return nil
// // 	}else if t.key == -5{
// // 		return -5
// // 	}
// // }
