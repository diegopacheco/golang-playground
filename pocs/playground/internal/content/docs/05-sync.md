# The sync Package

Channels are not the only tool. The `sync` package gives you classic primitives for shared memory: `WaitGroup`, `Mutex`, `RWMutex` and `Once`.

## When to use what

| Need | Tool |
|---|---|
| Wait for N goroutines to finish | `sync.WaitGroup` |
| Protect a shared value | `sync.Mutex` |
| Many readers, few writers | `sync.RWMutex` |
| Run setup exactly once | `sync.Once` / `sync.OnceValue` |
| Pass ownership of data | **channel** |

## Mutex

```go
package main

import (
	"fmt"
	"sync"
)

type Counter struct {
	mu sync.Mutex
	n  map[string]int
}

func (c *Counter) Inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n[key]++
}

func main() {
	c := Counter{n: map[string]int{}}
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() { c.Inc("hits") })
	}
	wg.Wait()
	fmt.Println("hits:", c.n["hits"])
}
```

Without the mutex the map would be written by many goroutines at once, which is a **data race** and Go's map would crash with `concurrent map writes`.

> [!TIP] Run your tests with `go test -race`. The race detector finds unsynchronized access at runtime.

## Once

```go
package main

import (
	"fmt"
	"sync"
)

var config = sync.OnceValue(func() string {
	fmt.Println("loading config...")
	return "prod"
})

func main() {
	for range 3 {
		fmt.Println(config())
	}
}
```

"loading config..." prints only once.

> [!WARNING] Never copy a `Mutex` or `WaitGroup` after first use. Pass a pointer to the struct that holds it. `go vet` catches this with its `copylocks` check.

## Key takeaways

- `WaitGroup` waits, `Mutex` protects, `Once` initializes.
- Keep the lock scope small and use `defer Unlock()`.
- Prefer channels for ownership transfer and mutexes for guarding state.
