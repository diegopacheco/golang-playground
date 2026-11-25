package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

func task(ctx context.Context, id int) error {
	select {
	case <-time.After(time.Duration(id) * 20 * time.Millisecond):
		if id == 2 {
			return errors.New("task 2 failed")
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	var wg sync.WaitGroup
	for id := 1; id <= 4; id++ {
		wg.Go(func() {
			if err := task(ctx, id); err != nil {
				cancel(err)
				fmt.Printf("task %d: %v\n", id, err)
				return
			}
			fmt.Printf("task %d: ok\n", id)
		})
	}
	wg.Wait()
	fmt.Println("first error:", context.Cause(ctx))
}
