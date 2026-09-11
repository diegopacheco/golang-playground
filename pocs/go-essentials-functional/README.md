# go-essentials-functional

Functional programming in Go with generics: generic `Map`/`Filter`/`Reduce`, closures that capture state, and function composition.

### How it works

`src/main.go` defines generic `Map`, `Filter`, and `Reduce` over any slice, builds closures (`makeCounter`, `multiplier`) that capture and mutate state, and composes two functions into a single pipeline with a generic `compose`.

### Run

```bash
./run.sh
```

### Output

```
map/filter/reduce: [2 4 6 8 10 12] [2 4 6] 21
closure counter: 1 2 3
closure multiplier: 30
compose (x+1)*2 of 5: 12
```
