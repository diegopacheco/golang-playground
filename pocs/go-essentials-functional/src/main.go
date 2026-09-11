package main

import "fmt"

func Map[T, R any](in []T, f func(T) R) []R {
	out := make([]R, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

func Filter[T any](in []T, keep func(T) bool) []T {
	out := []T{}
	for _, v := range in {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

func Reduce[T, R any](in []T, acc R, f func(R, T) R) R {
	for _, v := range in {
		acc = f(acc, v)
	}
	return acc
}

func makeCounter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

func multiplier(factor int) func(int) int {
	return func(x int) int { return x * factor }
}

func compose[T any](f, g func(T) T) func(T) T {
	return func(x T) T { return f(g(x)) }
}

func main() {
	nums := []int{1, 2, 3, 4, 5, 6}
	doubled := Map(nums, func(x int) int { return x * 2 })
	evens := Filter(nums, func(x int) bool { return x%2 == 0 })
	sum := Reduce(nums, 0, func(a, x int) int { return a + x })
	fmt.Println("map/filter/reduce:", doubled, evens, sum)

	counter := makeCounter()
	fmt.Println("closure counter:", counter(), counter(), counter())

	triple := multiplier(3)
	fmt.Println("closure multiplier:", triple(10))

	incThenDouble := compose(func(x int) int { return x * 2 }, func(x int) int { return x + 1 })
	fmt.Println("compose (x+1)*2 of 5:", incThenDouble(5))
}
