package main

import (
	"fmt"
	"sort"
	"time"
)

func generator(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

func fanIn(a, b <-chan int) []int {
	out := []int{}
	for a != nil || b != nil {
		select {
		case v, ok := <-a:
			if !ok {
				a = nil
				continue
			}
			out = append(out, v)
		case v, ok := <-b:
			if !ok {
				b = nil
				continue
			}
			out = append(out, v)
		}
	}
	sort.Ints(out)
	return out
}

func withTimeout() string {
	slow := make(chan int)
	go func() {
		time.Sleep(50 * time.Millisecond)
		slow <- 1
	}()
	select {
	case <-slow:
		return "got value"
	case <-time.After(10 * time.Millisecond):
		return "timeout"
	}
}

func main() {
	pipeline := square(generator(1, 2, 3, 4))
	piped := []int{}
	for v := range pipeline {
		piped = append(piped, v)
	}
	fmt.Println("pipeline gen->square:", piped)

	fmt.Println("fanIn:", fanIn(generator(1, 3, 5), generator(2, 4, 6)))
	fmt.Println("withTimeout:", withTimeout())
}
