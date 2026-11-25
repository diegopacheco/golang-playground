package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	start := time.Now()
	for i := 1; i <= 5; i++ {
		wg.Go(func() {
			time.Sleep(time.Duration(i) * 10 * time.Millisecond)
			fmt.Printf("goroutine %d finished\n", i)
		})
	}
	wg.Wait()
	fmt.Printf("5 goroutines ran concurrently in ~%dms, not 150ms\n", time.Since(start).Milliseconds()/10*10)
}
