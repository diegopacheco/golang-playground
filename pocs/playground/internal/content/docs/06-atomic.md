# Atomics

`sync/atomic` gives lock free operations on single values. They are the fastest way to share a counter or a flag.

## Typed atomics

```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var ops atomic.Int64
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			for range 1000 {
				ops.Add(1)
			}
		})
	}
	wg.Wait()
	fmt.Println("ops:", ops.Load())
}
```

| Type | Useful methods |
|---|---|
| `atomic.Int64`, `atomic.Int32` | `Add`, `Load`, `Store`, `CompareAndSwap` |
| `atomic.Bool` | `Load`, `Store`, `Swap` |
| `atomic.Pointer[T]` | `Load`, `Store` for copy on write config |
| `atomic.Value` | any value, same type every time |

## Compare and swap

`CompareAndSwap(old, new)` sets the value only when it still equals `old`. It is the building block of lock free algorithms.

> [!WARNING] Atomics protect **one** value. When two values must change together, use a `Mutex`.

## Key takeaways

- Use typed atomics such as `atomic.Int64` instead of the old functions.
- Perfect for counters, flags and swapping a config pointer.
- Anything bigger than one value needs a mutex.
