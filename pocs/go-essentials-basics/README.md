# go-essentials-basics

Go building blocks: `for` loops, `for` as while, ranging over collections, maps, `switch`, `sort.Slice`, structs with methods, and struct embedding.

### How it works

`src/main.go` has one function per concept. `forLoop` sums a range, `whileLoop` halves a number counting steps, `loopOverCollection` ranges with index, `mapsAndSwitch` combines a map lookup with a tagless `switch`, `sorting` sorts a slice of structs by different keys, and `Dog` embeds `Animal` to reuse its `Describe` method and fields.

### Run

```bash
./run.sh
```

### Output

```
forLoop sum 1..5: 15
whileLoop halving steps: 3
loopOverCollection: [0:apple 1:banana 2:cherry]
mapsAndSwitch(2): two
mapsAndSwitch(9): many
sorted by age: [{bob 25} {alice 30} {carol 35}]
sorted by name desc: [{carol 35} {bob 25} {alice 30}]
struct describe: Rex has 4 legs
struct method: woof
embedded field: Rex
```
