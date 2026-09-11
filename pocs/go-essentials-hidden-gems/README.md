# go-essentials-hidden-gems

Lesser-known Go features: labeled `break`, `iota` for enums and bit shifts, named return values, `defer` running LIFO, the `maps` and `slices` helpers, and `strings.Builder`.

### How it works

`src/main.go` breaks out of nested loops with a label, builds `KB`/`MB`/`GB` constants with `iota` and bit shifts, uses named returns, proves `defer` runs in last-in-first-out order, sorts map keys with `slices.Sorted(maps.Keys(...))` and clones a map with `maps.Clone`, and builds a string efficiently with `strings.Builder`.

### Run

```bash
./run.sh
```

### Output

```
labeled break: 1+2
iota sizes: KB=1024 MB=1048576 GB=1073741824
named returns: 17/5=3 rem=2
defer runs LIFO: [2 1 0]
sorted map keys: [a b c] cloned: map[a:1 b:2 c:3 d:4]
strings.Builder: [0][1][2]
```
