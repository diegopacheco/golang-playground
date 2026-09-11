package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed on %q: %s", e.Field, e.Msg)
}

func lookup(id int) error {
	if id < 0 {
		return fmt.Errorf("lookup id=%d: %w", id, ErrNotFound)
	}
	return nil
}

func validate(age int) error {
	if age < 0 {
		return &ValidationError{Field: "age", Msg: "must be non-negative"}
	}
	return nil
}

func main() {
	err := lookup(-1)
	fmt.Println("wrapped error:", err)
	fmt.Println("errors.Is ErrNotFound:", errors.Is(err, ErrNotFound))

	verr := validate(-5)
	var ve *ValidationError
	if errors.As(verr, &ve) {
		fmt.Printf("errors.As field=%s msg=%s\n", ve.Field, ve.Msg)
	}

	joined := errors.Join(ErrNotFound, &ValidationError{Field: "name", Msg: "required"})
	fmt.Println("errors.Join:", joined)
	fmt.Println("joined Is ErrNotFound:", errors.Is(joined, ErrNotFound))

	fmt.Println("lookup(1) ok:", lookup(1) == nil)
}
