# Generics

Type parameters let you write one function or type for many types, checked at compile time.

## Generic functions

```go
package main

import (
	"cmp"
	"fmt"
)

func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

func Max[T cmp.Ordered](xs ...T) T {
	m := xs[0]
	for _, x := range xs[1:] {
		m = max(m, x)
	}
	return m
}

func main() {
	fmt.Println(Map([]int{1, 2, 3}, func(i int) string { return fmt.Sprint(i * 10) }))
	fmt.Println(Max(3, 9, 4), Max("go", "rust", "c"))
}
```

## Constraints

| Constraint | Allows |
|---|---|
| `any` | every type |
| `comparable` | types usable with `==`, such as map keys |
| `cmp.Ordered` | types usable with `<` |
| `~int | ~float64` | any type whose underlying type is listed |

## A generic channel helper

```go
package main

import "fmt"

func FromSlice[T any](xs []T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for _, x := range xs {
			out <- x
		}
	}()
	return out
}

func main() {
	for s := range FromSlice([]string{"a", "b", "c"}) {
		fmt.Println(s)
	}
}
```

> [!TIP] Reach for generics when you would otherwise copy the same code for several types. For one type, plain code is clearer.

## Key takeaways

- `[T any]` declares a type parameter.
- Constraints say what operations `T` supports.
- Generics combine nicely with channels for reusable concurrency helpers.
