package main

import (
	"cmp"
	"fmt"
)

type Number interface {
	~int | ~int64 | ~float64
}

func Sum[T Number](in []T) T {
	var total T
	for _, v := range in {
		total += v
	}
	return total
}

func Max[T cmp.Ordered](in []T) T {
	best := in[0]
	for _, v := range in[1:] {
		if v > best {
			best = v
		}
	}
	return best
}

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

type Celsius float64

func Keys[K comparable, V any](m map[K]V) int {
	return len(m)
}

func main() {
	fmt.Println("Sum ints:", Sum([]int{1, 2, 3, 4}))
	fmt.Println("Sum floats:", Sum([]float64{1.5, 2.5}))
	fmt.Println("Sum defined type:", Sum([]Celsius{10, 20, 30}))
	fmt.Println("Max:", Max([]string{"apple", "pear", "kiwi"}))

	s := &Stack[string]{}
	s.Push("a")
	s.Push("b")
	top, _ := s.Pop()
	fmt.Println("Stack pop:", top)

	fmt.Println("Keys count:", Keys(map[string]int{"x": 1, "y": 2}))
}
