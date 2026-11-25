# Defer, Panic and Recover

`defer` schedules cleanup, `panic` aborts the current goroutine and `recover` stops a panic inside a deferred call.

## Defer

Deferred calls run when the function returns, in **LIFO** order.

```go
package main

import "fmt"

func main() {
	for i := range 3 {
		defer fmt.Println("deferred", i)
	}
	fmt.Println("body done")
}
```

Use it for `Unlock`, `Close` and `cancel` right after you acquire the resource.

## Recover

```go
package main

import (
	"errors"
	"fmt"
)

func safeDiv(a, b int) (q int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint("recovered: ", r))
		}
	}()
	return a / b, nil
}

func main() {
	fmt.Println(safeDiv(10, 2))
	fmt.Println(safeDiv(1, 0))
}
```

> [!WARNING] A panic in **any** goroutine that is not recovered crashes the **whole program**. `recover` only works in the same goroutine, so a server must recover inside every goroutine it starts.

> [!TIP] Return `error` for expected failures. Keep `panic` for bugs and impossible states.

## Key takeaways

- `defer` runs on return, last in first out.
- `recover` works only inside a deferred function.
- Errors are values: prefer returning them over panicking.
