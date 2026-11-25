# Pipelines and Fan-out / Fan-in

A pipeline is a series of stages connected by channels. Each stage is a group of goroutines that receive values, transform them and send them on.

## Stages

```go
package main

import (
	"fmt"
	"sync"
)

func gen(nums ...int) <-chan int {
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

func merge(cs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	for _, c := range cs {
		wg.Go(func() {
			for v := range c {
				out <- v
			}
		})
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {
	src := gen(1, 2, 3, 4, 5, 6)
	a, b := square(src), square(src)
	sum := 0
	for v := range merge(a, b) {
		sum += v
	}
	fmt.Println("sum of squares:", sum)
}
```

## Fan-out and fan-in

- **Fan-out**: several stages read from the **same** channel to spread the work, like `a` and `b` above.
- **Fan-in**: `merge` joins several channels into one.

## Rules for a healthy pipeline

1. Each stage **owns and closes** its output channel.
2. Each stage ranges over its input until it is closed.
3. Add a `ctx` or `done` channel so a consumer can stop early without leaking upstream goroutines.

> [!TIP] Go 1.23 range over func iterators (`iter.Seq`) are a lighter choice for pipelines that do **not** need concurrency.

## Key takeaways

- Stages are functions that take `<-chan` and return `<-chan`.
- Fan-out spreads work, fan-in merges results.
- Closing flows downstream: when the source closes, every stage drains and closes.
