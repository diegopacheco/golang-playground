# go-essentials-context

The `context` package: cancellation, timeouts, deadlines, and request-scoped values.

### How it works

`src/main.go` cancels a worker with `context.WithCancel`, races work against `context.WithTimeout`, checks `context.DeadlineExceeded` via `errors.Is` with `context.WithDeadline`, and carries a value through `context.WithValue` using a typed key.

### Run

```bash
./run.sh
```

### Output

```
withCancel: worker stopped: context canceled
withTimeout: cancelled: context deadline exceeded
withDeadline exceeded: true
withValue: alice
```
