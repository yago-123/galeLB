package main

import (
	"fmt"

	"github.com/yago-123/gale/ring"
)

func main() {
	// Create consistent hashing with 3 virtual nodes per real node
	ch, err := ring.New(5)
	if err != nil {
		panic(err)
	}

	// Add some nodes
	ch.AddNode("Node1")
	ch.AddNode("Node2")
	ch.AddNode("Node3")

	// Hash the IP of a request
	for i := 0; i < 15; i++ {
		fmt.Printf("Request from IP will be routed to %s\n", ch.GetNode(fmt.Sprintf("193.168.1.%d", i)))
	}
}
