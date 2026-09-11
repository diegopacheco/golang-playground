# go-essentials-interfaces

Interfaces in Go: satisfying an interface implicitly, type assertions, type switches, and interface embedding.

### How it works

`src/main.go` defines a `Shape` interface implemented by `Rectangle` and `Circle`, uses a type assertion to reach the optional `Stringer`, branches on a value's dynamic type with a type `switch`, and composes `Reader` and `Closer` into an embedded `ReadCloser` interface.

### Run

```bash
./run.sh
```

### Output

```
describe: area=12.00 perimeter=14.00
describe: area=78.54 perimeter=31.42
type assertion: Circle(r=2)
typeSwitch int: int:42
typeSwitch string: string:hi
typeSwitch shape: shape
embedded interface: data from log.txt | closed log.txt
```
