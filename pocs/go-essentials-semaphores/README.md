# go-essentials-semaphores

Concurrency limiting and synchronization: a counting semaphore built from a buffered channel, `sync.RWMutex` for reader/writer access, and `sync/atomic` for lock-free counters.

### How it works

`src/main.go` caps concurrency at 3 with a buffered channel used as a semaphore and tracks the peak in-flight count with atomics to prove the cap holds. `rwMutex` runs concurrent readers under `RLock` alongside a single `Lock` writer, and `waitGroupCollect` sums work from 10 goroutines with `atomic.AddInt64`.

### Run

```bash
./run.sh
```

### Output

```
boundedConcurrency max in-flight (limit 3): 3
rwMutex data=42 reads=5
waitGroupCollect sum 1..10: 55
```
