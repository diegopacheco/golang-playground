# go-essentials-gorm

The GORM ORM against a pure-Go in-memory SQLite (`github.com/glebarez/sqlite`, no cgo): auto-migration, associations with `Preload`, and CRUD. Runs with no external service.

### How it works

`src/main.go` opens an in-memory SQLite DB, auto-migrates `Author` and `Book` (a has-many association), creates an author with nested books in one call, eager-loads the books with `Preload`, updates and deletes rows, and reads a single column with `Pluck`.

### Run

```bash
./run.sh
```

### Output

```
authors count: 2
author "Rob Pike" has 2 books
  book: The Go Programming Language
  book: The Practice of Programming
updated book: The Go Programming Language (2nd)
authors after delete: 1
all book titles: [The Go Programming Language (2nd) The Practice of Programming]
```
