package main

import "fmt"

func main() {
	recruse(3, 3, 40)
}

var breath = 0

func dive(msgs []int, leaders int, depth int, limit int) {
	depth++
	if depth > limit {
		fmt.Println("Breath ", breath)
		breath++
		return
	}

	for d, v := range msgs {
		msgs2 := append(msgs[0:d], msgs[d+1:]...)
		ml2 := len(msgs2)
		for i := 0; i < leaders-1; i++ {
			msgs2 = append(msgs2, v)
			dive(msgs2, leaders, depth, limit)
			msgs2 = msgs2[:ml2]
		}
	}
}

func recurse(msg int, leaders int, limit int) {
	var msgs []int
	for i := 0; i < msg; i++ {
		for j := 0; j < leaders-1; j++ {
			msgs = append(msgs, i*1024+j)
		}
	}
	dive(msgs, leaders, 0, limit)
}
