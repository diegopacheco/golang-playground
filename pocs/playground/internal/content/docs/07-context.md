# Context

`context.Context` carries cancellation, deadlines and request scoped values across API boundaries and goroutines. It is how you **stop** goroutines.

## The tree of contexts

Contexts form a tree. Cancelling a parent cancels **every** child.

| Constructor | Use it for |
|---|---|
| `context.Background()` | the root, in `main` and tests |
| `context.WithCancel(parent)` | manual cancellation |
| `context.WithTimeout(parent, d)` | cancel after a duration |
| `context.WithDeadline(parent, t)` | cancel at a point in time |
| `context.WithValue(parent, k, v)` | request scoped data such as a correlation id |

## Stopping a worker

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func worker(ctx context.Context, out chan<- int) {
	defer close(out)
	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return
		case out <- i:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Millisecond)
	defer cancel()
	out := make(chan int)
	go worker(ctx, out)
	for v := range out {
		fmt.Println("tick", v)
	}
	fmt.Println("stopped:", ctx.Err())
}
```

## Rules

1. Pass `ctx` as the **first parameter**: `func Fetch(ctx context.Context, id string)`.
2. Always `defer cancel()` to release resources.
3. Never store a context in a struct, pass it along.
4. Use `WithValue` only for request scoped data, never for optional parameters.

> [!TIP] `ctx.Err()` tells you **why** it ended: `context.Canceled` or `context.DeadlineExceeded`. `context.Cause(ctx)` gives the custom cause set with `WithCancelCause`.

## Key takeaways

- Context is the standard way to cancel work.
- Every blocking select in a goroutine should have a `case <-ctx.Done()`.
- Cancelling a parent cancels all of its children.
