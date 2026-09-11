# go-essentials-generics

Go generics (type parameters): custom constraints, the `cmp.Ordered` constraint, a generic container type, and `comparable` map keys.

### How it works

`src/main.go` defines a `Number` constraint with a type union (including `~` for defined types like `Celsius`), a `Sum` over it, a `Max` over `cmp.Ordered`, a generic `Stack[T]` with `Push`/`Pop`, and a `Keys` function constrained to `comparable` keys.

### Run

```bash
./run.sh
```

### Output

```
Sum ints: 10
Sum floats: 4
Sum defined type: 60
Max: pear
Stack pop: b
Keys count: 2
```
