# Channels

Channels are typed pipes that connect goroutines. Send a value in one goroutine and receive it in another, with synchronization built in.

## The philosophy

> [!INFO] **"Do not communicate by sharing memory; instead, share memory by communicating."** That is the Go proverb behind channels.

Instead of many goroutines touching the same variable behind a lock, one goroutine **owns** the data and passes it over a channel.

## Basics

```go
package main

import "fmt"

func main() {
	ch := make(chan string)
	go func() {
		ch <- "ping"
	}()
	msg := <-ch
	fmt.Println(msg)
}
```

| Operation | Syntax | Blocks until |
|---|---|---|
| Create | `make(chan T)` | - |
| Send | `ch <- v` | a receiver is ready (unbuffered) |
| Receive | `v := <-ch` | a value is available |
| Receive with status | `v, ok := <-ch` | a value arrives or the channel is closed |
| Close | `close(ch)` | never blocks |

## Unbuffered channels are a handshake

An unbuffered channel has no storage. A send blocks until a receiver takes the value, so the two goroutines **meet** at that point. This is called a *rendezvous* and it is a synchronization point as well as a data transfer.

## Closing and ranging

Close a channel to say "no more values". Receivers can loop with `range` until it is closed:

```go
package main

import "fmt"

func produce(n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := range n {
			out <- i * i
		}
	}()
	return out
}

func main() {
	for v := range produce(5) {
		fmt.Println(v)
	}
}
```

## Directional channels

Function signatures can limit what a caller may do:

- `chan<- int` is **send only**.
- `<-chan int` is **receive only**.

The compiler then rejects a receive on a send only channel, which documents ownership in the type.

## The rules of closing

| Action | nil channel | open channel | closed channel |
|---|---|---|---|
| Send | blocks forever | blocks or sends | **panic** |
| Receive | blocks forever | blocks or receives | zero value, `ok == false` |
| Close | **panic** | closes | **panic** |

> [!WARNING] Only the **sender** closes a channel, and only once. Closing from the receiver side or closing twice causes a panic.

> [!TIP] You do not have to close every channel. Close it only when receivers need to know that no more values will come, such as when they use `range`.

## Key takeaways

- Channels are typed, goroutine safe pipes.
- Unbuffered channels synchronize sender and receiver.
- The sender closes; receivers `range` until close.
- Use directional types to make ownership clear.
