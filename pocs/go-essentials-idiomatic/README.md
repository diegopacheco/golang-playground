# go-essentials-idiomatic

Idiomatic Go patterns: the functional-options constructor, the comma-ok idiom, `defer` for cleanup, and zero values.

### How it works

`src/main.go` builds a `Server` with variadic functional options (`WithHost`, `WithPort`, `WithTLS`, `WithTimeout`) over sensible defaults, reads a map with the comma-ok idiom, shows `defer` releasing resources in reverse acquisition order, and prints the zero values every Go type gets for free.

### Run

```bash
./run.sh
```

### Output

```
default: http://localhost:8080 timeout=30s
configured: https://api.local:443 timeout=10s
comma-ok: found 1
defer cleanup: acquire db -> acquire lock -> work -> release lock -> release db
zero values: int=0 str="" bool=false slice=[] nil=true
```
