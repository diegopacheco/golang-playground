package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

func labeledBreak() string {
	found := ""
outer:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if i+j == 3 {
				found = fmt.Sprintf("%d+%d", i, j)
				break outer
			}
		}
	}
	return found
}

func iotaEnum() string {
	const (
		_ = iota
		KB = 1 << (10 * iota)
		MB
		GB
	)
	return fmt.Sprintf("KB=%d MB=%d GB=%d", KB, MB, GB)
}

func multiReturn() (quotient, remainder int) {
	quotient = 17 / 5
	remainder = 17 % 5
	return
}

func deferStack() []int {
	out := []int{}
	func() {
		for i := 0; i < 3; i++ {
			defer func(n int) { out = append(out, n) }(i)
		}
	}()
	return out
}

func mapsAndSlices() ([]string, map[string]int) {
	m := map[string]int{"c": 3, "a": 1, "b": 2}
	keys := slices.Sorted(maps.Keys(m))
	cloned := maps.Clone(m)
	cloned["d"] = 4
	return keys, cloned
}

func stringBuilder() string {
	var b strings.Builder
	for i := 0; i < 3; i++ {
		fmt.Fprintf(&b, "[%d]", i)
	}
	return b.String()
}

func main() {
	fmt.Println("labeled break:", labeledBreak())
	fmt.Println("iota sizes:", iotaEnum())
	q, r := multiReturn()
	fmt.Printf("named returns: 17/5=%d rem=%d\n", q, r)
	fmt.Println("defer runs LIFO:", deferStack())
	keys, cloned := mapsAndSlices()
	fmt.Println("sorted map keys:", keys, "cloned:", cloned)
	fmt.Println("strings.Builder:", stringBuilder())
}
