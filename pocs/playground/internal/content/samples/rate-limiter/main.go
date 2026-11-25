package main

import (
	"fmt"
	"time"
)

func main() {
	requests := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		requests <- i
	}
	close(requests)
	limiter := time.NewTicker(20 * time.Millisecond)
	defer limiter.Stop()
	start := time.Now()
	for req := range requests {
		<-limiter.C
		fmt.Printf("request %d at +%dms\n", req, time.Since(start).Milliseconds()/10*10)
	}
}
