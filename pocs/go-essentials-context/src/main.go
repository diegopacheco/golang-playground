package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func withCancel() string {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan string)
	go func() {
		<-ctx.Done()
		done <- "worker stopped: " + ctx.Err().Error()
	}()
	cancel()
	return <-done
}

func withTimeout() string {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	select {
	case <-time.After(100 * time.Millisecond):
		return "work finished"
	case <-ctx.Done():
		return "cancelled: " + ctx.Err().Error()
	}
}

func withDeadline() bool {
	deadline := time.Now().Add(10 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	<-ctx.Done()
	return errors.Is(ctx.Err(), context.DeadlineExceeded)
}

type key string

func withValue() string {
	ctx := context.WithValue(context.Background(), key("user"), "alice")
	return ctx.Value(key("user")).(string)
}

func main() {
	fmt.Println("withCancel:", withCancel())
	fmt.Println("withTimeout:", withTimeout())
	fmt.Println("withDeadline exceeded:", withDeadline())
	fmt.Println("withValue:", withValue())
}
