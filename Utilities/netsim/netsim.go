package main

import (
	"fmt"
	"math/rand"
)

type Node struct {
	id          int
	MsgSeen     int
	MsgSent     bool
	Connections []*Node
}

const (
	NumNodes    = 7000 // Number of nodes in sim
	Connections = 400  // Percent connected to
	Broadcast   = 2    // Who you broadcast to
)

var Nodes []*Node

func Stats() (seen int, sent int) {
	for _, n := range Nodes {
		if n.MsgSeen > 0 {
			seen++
		}
		if n.MsgSent {
			sent++
		}
	}
	return
}

func Member(cs []*Node, id int) bool {
	for _, n := range cs {
		if n.id == id {
			return true
		}
	}
	return false
}

func OneTest() {
	Nodes = Nodes[:0]
	for i := 0; i < NumNodes; i++ {
		n := new(Node)
		n.id = i
		Nodes = append(Nodes, n)
	}

	for i := 0; i < NumNodes; i++ {
		for j := 0; j < Connections; j++ {
			c := rand.Intn(len(Nodes))
			for c == i || Member(Nodes[i].Connections, c) {
				c = rand.Intn(len(Nodes))
			}
			Nodes[i].Connections = append(Nodes[i].Connections, Nodes[c])
		}
	}
	if true {
		for _, n1 := range Nodes {
			for _, n2 := range n1.Connections {
				if !Member(n2.Connections, n1.id) {
					n2.Connections = append(n2.Connections, n1)
				}
			}
		}
	}
	for i := 0; false && i < NumNodes; i++ {
		fmt.Print("node ", i, " Connections [")
		for _, n := range Nodes[i].Connections {
			fmt.Print(n.id, " ")
		}
		fmt.Println("]")
	}

	Nodes[0].MsgSeen = 1

	collide := 0
	for step := 2; true; step++ {
		var broadcasting []int
		var reaching []int
		var collisions []int

		seen, sent := Stats()

		for i, n := range Nodes {
			if n.MsgSeen > 0 && n.MsgSeen < step && !n.MsgSent {
				broadcasting = append(broadcasting, i)
				// Broadcast our message
				start := rand.Intn(len(Nodes[i].Connections))
				var sentto []int
				for i := 0; i < Broadcast; i++ {
					c := start
					for {
						c = rand.Intn(len(Nodes[i].Connections))
						for _, v := range sentto {
							if v == c {
								continue
							}
						}
						sentto = append(sentto, c)
						break
					}
					//c = (start + i) % len(Nodes[i].Connections)
					node := Nodes[i].Connections[c]
					if node.MsgSeen == 0 {
						reaching = append(reaching, node.id)
						node.MsgSeen = step
					} else {
						collide++
						collisions = append(collisions, node.id)
					}
				}
				n.MsgSent = true
			}
		}

		fmt.Printf("Step %4d, Collide %4d Seen %4d Sent %4d %v %v %v\n", step-1, collide, seen, sent, broadcasting, reaching, collisions)

		if len(broadcasting) == 0 {
			seen, sent := Stats()
			fmt.Printf("DONE %4d, Collide %4d Seen %4d Sent %4d \n", step-1, collide, seen, sent)
			return
		}
	}

}

func main() {
	for i := 0; i < 10; i++ {
		OneTest()
	}
}
