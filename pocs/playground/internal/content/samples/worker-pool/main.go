package main

import (
	"fmt"
	"sort"
	"sync"
)

func main() {
	jobs := make(chan int)
	results := make(chan string)
	var wg sync.WaitGroup
	for w := 1; w <= 3; w++ {
		wg.Go(func() {
			for n := range jobs {
				results <- fmt.Sprintf("%2d squared is %3d", n, n*n)
			}
		})
	}
	go func() {
		for n := 1; n <= 10; n++ {
			jobs <- n
		}
		close(jobs)
	}()
	go func() {
		wg.Wait()
		close(results)
	}()
	var out []string
	for r := range results {
		out = append(out, r)
	}
	sort.Strings(out)
	for _, line := range out {
		fmt.Println(line)
	}
}
