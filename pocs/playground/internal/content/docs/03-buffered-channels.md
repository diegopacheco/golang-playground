# Buffered Channels

A buffered channel has a queue of fixed capacity. Senders only block when the buffer is full, and receivers only block when it is empty.

## Creating one

```go
ch := make(chan int, 3)
```

`len(ch)` is the number of queued values and `cap(ch)` is the capacity.

```go
package main

import "fmt"

func main() {
	ch := make(chan string, 2)
	ch <- "first"
	ch <- "second"
	fmt.Println("len:", len(ch), "cap:", cap(ch))
	fmt.Println(<-ch)
	fmt.Println(<-ch)
}
```

No goroutine is needed here because both sends fit in the buffer.

## Unbuffered vs buffered

| | Unbuffered `make(chan T)` | Buffered `make(chan T, n)` |
|---|---|---|
| Send blocks when | no receiver is waiting | buffer is full |
| Receive blocks when | no sender is waiting | buffer is empty |
| Guarantee | the receiver **got** the value | the value is **queued** |
| Use it for | synchronization and handoff | smoothing bursts, limiting concurrency |

## Semaphore pattern

A buffered channel of empty structs limits how many goroutines do something at once:

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	sem := make(chan struct{}, 2)
	var wg sync.WaitGroup
	for i := range 6 {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			fmt.Println("job", i, "running")
			time.Sleep(20 * time.Millisecond)
		})
	}
	wg.Wait()
	fmt.Println("done, never more than 2 at once")
}
```

> [!WARNING] A buffer does **not** fix a deadlock, it only delays it. If nobody ever receives, the sender blocks once the buffer is full.

> [!TIP] Choose the capacity on purpose. `1` is great for "latest signal" patterns, `n` is great for a semaphore of `n` slots. A huge buffer usually hides a design problem.

## Key takeaways

- `make(chan T, n)` creates a queue of `n` slots.
- Sends block only when full, receives only when empty.
- A buffered channel of `struct{}` is a simple semaphore.
