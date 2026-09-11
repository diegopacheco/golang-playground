# go-essentials-goroutines

Goroutines every way: raw goroutines with `sync.WaitGroup`, a `sync.Mutex` for shared state, a channel-based worker pool, and `sync.Once` for one-time init.

### How it works

`src/main.go` runs one function per style. `goroutineWithMutex` proves a lock stops races (4 goroutines × 10000 increments == 40000), `workerPool` fans work across 3 workers reading from a jobs channel and writing to a results channel, and `sync.Once` guarantees the init runs exactly once even when called in a loop.

### Run

```bash
./run.sh
```

### Output

```
basicGoroutines: [0 1 4 9 16]
goroutineWithMutex: 40000
workerPool: [0 1 4 9 16 25 36 49]
onceInit: initialized
```
