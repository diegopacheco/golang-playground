package main

import (
	"fmt"
	"time"
)

func fetch(name string, delay time.Duration) <-chan string {
	ch := make(chan string, 1)
	go func() {
		time.Sleep(delay)
		ch <- name
	}()
	return ch
}

func main() {
	fast := fetch("fast-service", 20*time.Millisecond)
	slow := fetch("slow-service", 300*time.Millisecond)
	timeout := time.After(100 * time.Millisecond)
	for range 2 {
		select {
		case r := <-fast:
			fmt.Println("got", r)
		case r := <-slow:
			fmt.Println("got", r)
		case <-timeout:
			fmt.Println("timeout reached, slow-service ignored")
			return
		}
	}
}
