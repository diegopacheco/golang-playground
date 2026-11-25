# Goroutines

Goroutines are Go's lightweight concurrent functions, often called coroutines. They are cheap to start, the runtime schedules them, and they are what makes Go concurrency feel simple.

## What is a goroutine?

A goroutine is a function that runs **concurrently** with the rest of the program. Put the `go` keyword in front of any function call and it starts right away while the caller keeps going.

| Property | OS thread | Goroutine |
|---|---|---|
| Initial stack | ~1-8 MB fixed | ~2 KB, grows on demand |
| Created by | the kernel | the Go runtime |
| Switch cost | kernel context switch | user space, a few hundred ns |
| Typical count | thousands | **millions** |

> [!INFO] The Go scheduler is an **M:N scheduler**: it runs M goroutines on N OS threads. `GOMAXPROCS` sets how many threads run Go code at once, and it defaults to the number of CPU cores.

## Your first goroutine

```go
package main

import (
	"fmt"
	"time"
)

func say(msg string) {
	for i := range 3 {
		fmt.Println(msg, i)
		time.Sleep(10 * time.Millisecond)
	}
}

func main() {
	go say("goroutine")
	say("main")
}
```

The output interleaves because both functions run at the same time.

> [!WARNING] When `main` returns, the program exits and **every goroutine is killed**, finished or not. Never use `time.Sleep` to wait for goroutines. Use a `sync.WaitGroup` or a channel.

## Waiting the right way

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	for i := range 5 {
		wg.Go(func() {
			fmt.Println("worker", i, "done")
		})
	}
	wg.Wait()
	fmt.Println("all workers finished")
}
```

`wg.Go` (Go 1.25+) does the `Add(1)`, the `go` and the `defer Done()` for you.

## Loop variables

Since Go 1.22 every loop iteration gets **its own copy** of the loop variable, so capturing `i` inside a goroutine is safe. Older code passed it as a parameter: `go func(i int) { ... }(i)`.

## Goroutine leaks

A goroutine that blocks forever is never collected. This is a **leak**:

```go
func leak() {
	ch := make(chan int)
	go func() {
		ch <- 42
	}()
}
```

Nobody receives from `ch`, so the goroutine blocks on the send forever.

> [!TIP] Rule of thumb: **never start a goroutine without knowing how it stops.** Give every goroutine an exit through a closed channel or a cancelled `context.Context`.

## Key takeaways

- `go f()` starts `f` concurrently and returns at once.
- Goroutines are cheap: start thousands without worry.
- `main` returning kills everything, so always wait.
- Every goroutine needs a way to stop.
