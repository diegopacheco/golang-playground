# Select

`select` waits on several channel operations at once and runs the first one that is ready. It is the control center of concurrent Go code.

## Syntax

```go
select {
case v := <-a:
	use(v)
case b <- x:
	sent()
case <-time.After(time.Second):
	timeout()
default:
	nothingReady()
}
```

- When **several** cases are ready, Go picks one **at random**, which keeps it fair.
- When **none** is ready, `select` blocks, unless there is a `default`.

## Timeouts

```go
package main

import (
	"fmt"
	"time"
)

func slow() <-chan string {
	ch := make(chan string, 1)
	go func() {
		time.Sleep(200 * time.Millisecond)
		ch <- "slow result"
	}()
	return ch
}

func main() {
	select {
	case r := <-slow():
		fmt.Println(r)
	case <-time.After(50 * time.Millisecond):
		fmt.Println("timeout: gave up after 50ms")
	}
}
```

> [!TIP] The result channel is buffered with capacity 1, so the slow goroutine can still send and exit after the timeout. With an unbuffered channel it would leak.

## Non blocking operations

```go
select {
case msg := <-ch:
	fmt.Println("got", msg)
default:
	fmt.Println("no message, moving on")
}
```

## Multiplexing with a quit channel

```go
package main

import "fmt"

func main() {
	nums := make(chan int)
	quit := make(chan struct{})
	go func() {
		for i := range 3 {
			nums <- i
		}
		close(quit)
	}()
	for {
		select {
		case n := <-nums:
			fmt.Println("number", n)
		case <-quit:
			fmt.Println("quit")
			return
		}
	}
}
```

## Disabling a case with nil

A receive from a `nil` channel blocks forever, so setting a channel variable to `nil` **turns off** its case. This is the classic way to merge two channels until both are closed.

> [!WARNING] An empty `select {}` blocks forever. That is useful in a server main, and a bug everywhere else.

## Key takeaways

- `select` waits on many channels and runs one ready case.
- Random choice among ready cases keeps it fair.
- `default` makes it non blocking, `time.After` gives timeouts.
- A `nil` channel disables its case.
