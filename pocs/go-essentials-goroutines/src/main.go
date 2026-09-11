package main

import (
	"fmt"
	"sort"
	"sync"
)

func basicGoroutines() []int {
	var wg sync.WaitGroup
	results := make([]int, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			results[n] = n * n
		}(i)
	}
	wg.Wait()
	return results
}

func goroutineWithMutex() int {
	var mu sync.Mutex
	var wg sync.WaitGroup
	counter := 0
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10000; j++ {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return counter
}

func workerPool() []int {
	jobs := make(chan int, 8)
	results := make(chan int, 8)
	var wg sync.WaitGroup
	for w := 0; w < 3; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := range jobs {
				results <- n * n
			}
		}()
	}
	for i := 0; i < 8; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	close(results)
	out := []int{}
	for r := range results {
		out = append(out, r)
	}
	sort.Ints(out)
	return out
}

func onceInit() string {
	var once sync.Once
	value := ""
	for i := 0; i < 3; i++ {
		once.Do(func() { value = "initialized" })
	}
	return value
}

func main() {
	fmt.Println("basicGoroutines:", basicGoroutines())
	fmt.Println("goroutineWithMutex:", goroutineWithMutex())
	fmt.Println("workerPool:", workerPool())
	fmt.Println("onceInit:", onceInit())
}
