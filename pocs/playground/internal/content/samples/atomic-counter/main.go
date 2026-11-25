package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var requests atomic.Int64
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			for range 100 {
				requests.Add(1)
			}
		})
	}
	wg.Wait()
	fmt.Println("requests:", requests.Load())
}
