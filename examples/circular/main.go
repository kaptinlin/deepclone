// Package main demonstrates handling circular references with the deepclone library.
package main

import (
	"fmt"

	"github.com/kaptinlin/deepclone"
)

type node struct {
	ID   int
	Next *node
}

func main() {
	fmt.Println("=== Circular Reference Example ===")

	node1 := &node{ID: 1}
	node2 := &node{ID: 2}
	node1.Next = node2
	node2.Next = node1

	cloned := deepclone.Clone(node1)

	fmt.Printf("Original: %d -> %d -> %d\n",
		node1.ID, node1.Next.ID, node1.Next.Next.ID)
	fmt.Printf("Cloned: %d -> %d -> %d\n",
		cloned.ID, cloned.Next.ID, cloned.Next.Next.ID)
}
