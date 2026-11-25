package main

import (
	"fmt"
	"strings"
)

func Source[T any](xs ...T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for _, x := range xs {
			out <- x
		}
	}()
	return out
}

func Map[T, U any](in <-chan T, f func(T) U) <-chan U {
	out := make(chan U)
	go func() {
		defer close(out)
		for v := range in {
			out <- f(v)
		}
	}()
	return out
}

func main() {
	for s := range Map(Source("go", "channels", "rock"), strings.ToUpper) {
		fmt.Println(s)
	}
}
