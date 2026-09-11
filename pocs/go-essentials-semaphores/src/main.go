package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func boundedConcurrency() int32 {
	sem := make(chan struct{}, 3)
	var wg sync.WaitGroup
	var inFlight, maxSeen int32
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			cur := atomic.AddInt32(&inFlight, 1)
			for {
				old := atomic.LoadInt32(&maxSeen)
				if cur <= old || atomic.CompareAndSwapInt32(&maxSeen, old, cur) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			atomic.AddInt32(&inFlight, -1)
		}()
	}
	wg.Wait()
	return maxSeen
}

func rwMutex() (int, int) {
	var mu sync.RWMutex
	data := 0
	var wg sync.WaitGroup
	reads := 0
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.RLock()
			_ = data
			mu.RUnlock()
		}()
		reads++
	}
	mu.Lock()
	data = 42
	mu.Unlock()
	wg.Wait()
	return data, reads
}

func waitGroupCollect() int {
	var wg sync.WaitGroup
	var total int64
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			atomic.AddInt64(&total, int64(n))
		}(i)
	}
	wg.Wait()
	return int(total)
}

func main() {
	fmt.Println("boundedConcurrency max in-flight (limit 3):", boundedConcurrency())
	data, reads := rwMutex()
	fmt.Printf("rwMutex data=%d reads=%d\n", data, reads)
	fmt.Println("waitGroupCollect sum 1..10:", waitGroupCollect())
}
