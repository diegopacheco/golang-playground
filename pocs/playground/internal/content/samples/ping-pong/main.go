package main

import "fmt"

func player(name string, in <-chan int, out chan<- int, done chan<- struct{}) {
	for ball := range in {
		fmt.Println(name, "hits", ball)
		if ball == 6 {
			close(done)
			return
		}
		out <- ball + 1
	}
}

func main() {
	ping := make(chan int)
	pong := make(chan int)
	done := make(chan struct{})
	go player("ping", ping, pong, done)
	go player("pong", pong, ping, done)
	ping <- 1
	<-done
	fmt.Println("game over")
}
