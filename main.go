package main

import (
	"fmt"
	"time"
)

func preorder(node *Node) {
	if node != nil {
		fmt.Printf("%p => %+v\n", node, node)
		for i := 0; i < MAX_DEGREE+1; i++ {
			preorder(node.children[i])
		}
	}
}

func main() {
	var root Node
	var node *Node
	var n int = 1

	start := time.Now()
	node = Insert(&root, 1)
	timeElapsed := time.Since(start)
	fmt.Printf("( 1 , %v )\n", timeElapsed)

	for i := 1; i <= 1000; i++ {
		n = n + 10
		root = Node{}

		start = time.Now()
		node = Insert(&root, 1)
		for j := 2; j <= n; j++ {
			node = Insert(node, j)
		}
		timeElapsed = time.Since(start)

		// preorder(node)
		// fmt.Println()

		fmt.Printf("( %v , %v )\n", n, timeElapsed)
	}

	// fmt.Printf("%p\n", Search(node, 42))
	// node = Delete(node, 100)
	// node = Delete(node, 39)
	// preorder(node)
	// fmt.Println()
}
