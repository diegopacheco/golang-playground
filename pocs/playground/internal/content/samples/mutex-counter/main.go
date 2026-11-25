package main

import (
	"fmt"
	"sync"
)

type Stats struct {
	mu   sync.Mutex
	hits map[string]int
}

func (s *Stats) Hit(page string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hits[page]++
}

func main() {
	s := &Stats{hits: map[string]int{}}
	var wg sync.WaitGroup
	for i := range 300 {
		wg.Go(func() {
			s.Hit([]string{"/home", "/docs", "/play"}[i%3])
		})
	}
	wg.Wait()
	fmt.Println("/home:", s.hits["/home"], "/docs:", s.hits["/docs"], "/play:", s.hits["/play"])
}
