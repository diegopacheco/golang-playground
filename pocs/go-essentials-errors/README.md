# go-essentials-errors

Error handling in Go: sentinel errors, wrapping with `%w`, `errors.Is`, `errors.As` for custom error types, and `errors.Join`.

### How it works

`src/main.go` wraps a sentinel `ErrNotFound` with `fmt.Errorf("%w")` and unwraps it with `errors.Is`, defines a `ValidationError` type reached with `errors.As`, and combines errors with `errors.Join` while proving `Is` still matches inside the joined error.

### Run

```bash
./run.sh
```

### Output

```
wrapped error: lookup id=-1: not found
errors.Is ErrNotFound: true
errors.As field=age msg=must be non-negative
errors.Join: not found
validation failed on "name": required
joined Is ErrNotFound: true
lookup(1) ok: true
```
