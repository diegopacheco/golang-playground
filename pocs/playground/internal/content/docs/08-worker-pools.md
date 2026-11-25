# Worker Pools

A worker pool runs a fixed number of goroutines that pull jobs from a shared channel. It bounds concurrency and keeps CPU, memory and connections under control.

## Shape

- A **jobs** channel feeds work in.
- **N workers** range over jobs.
- A **results** channel collects output.
- A `WaitGroup` closes results when every worker is done.

## Full program

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type Job struct{ ID, N int }
type Result struct{ ID, Square, Worker int }

func worker(id int, jobs <-chan Job, results chan<- Result) {
	for j := range jobs {
		time.Sleep(5 * time.Millisecond)
		results <- Result{ID: j.ID, Square: j.N * j.N, Worker: id}
	}
}

func main() {
	jobs := make(chan Job)
	results := make(chan Result)
	var wg sync.WaitGroup
	for w := 1; w <= 3; w++ {
		wg.Go(func() { worker(w, jobs, results) })
	}
	go func() {
		for i := 1; i <= 9; i++ {
			jobs <- Job{ID: i, N: i}
		}
		close(jobs)
	}()
	go func() {
		wg.Wait()
		close(results)
	}()
	total := 0
	for r := range results {
		fmt.Printf("job %d -> %d (worker %d)\n", r.ID, r.Square, r.Worker)
		total += r.Square
	}
	fmt.Println("total:", total)
}
```

> [!INFO] Results arrive **out of order**. Carry the job ID in the result when the order matters, then sort or index by it.

## Sizing the pool

| Work type | Good starting size |
|---|---|
| CPU bound | `runtime.NumCPU()` |
| IO bound (HTTP, DB) | what the downstream can handle, often 10-100 |

> [!WARNING] Close `results` only after **all** workers finish, which is why a separate goroutine waits on the `WaitGroup`. Closing it early makes a worker panic on send.

## Key takeaways

- A fixed set of workers ranging over a jobs channel.
- `close(jobs)` tells workers to stop.
- A `WaitGroup` decides when results can be closed.
