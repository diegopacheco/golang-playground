# go-essentials-slices

The `slices` standard-library package plus classic slice tricks: sort, search, reverse, compact, insert, delete, stack pop, 2D slices, and the `Values` iterator.

### How it works

`src/main.go` exercises `slices.Sort`, `Contains`, `Index`, `Max`/`Min`, `Reverse`, `Compact`, `Insert`, `Delete`, and `Clone` (to avoid mutating the source), pops from a slice used as a stack, builds a nested 2D slice, and sums a slice through the `slices.Values` range-over-function iterator.

### Run

```bash
./run.sh
```

### Output

```
slices.Sort: [1 1 2 3 4 5 6 9]
slices.Contains 5: true
slices.Index 4: 2
slices.Max/Min: 9 1
slices.Reverse: [6 2 9 5 1 4 1 3]
slices.Compact: [1 2 3 4]
slices.Insert: [3 1 100 200 4 1 5 9 2 6]
slices.Delete: [3 1 5 9 2 6]
pop: 3 remaining: [1 2]
2d slice: [[0 1 2] [3 4 5]]
slices.Values sum: 60
```
