package main

import (
	"fmt"
	"slices"
)

func main() {
	nums := []int{3, 1, 4, 1, 5, 9, 2, 6}

	sorted := slices.Clone(nums)
	slices.Sort(sorted)
	fmt.Println("slices.Sort:", sorted)

	fmt.Println("slices.Contains 5:", slices.Contains(nums, 5))
	fmt.Println("slices.Index 4:", slices.Index(nums, 4))
	fmt.Println("slices.Max/Min:", slices.Max(nums), slices.Min(nums))

	reversed := slices.Clone(nums)
	slices.Reverse(reversed)
	fmt.Println("slices.Reverse:", reversed)

	withDup := []int{1, 1, 2, 3, 3, 3, 4}
	fmt.Println("slices.Compact:", slices.Compact(slices.Clone(withDup)))

	inserted := slices.Insert(slices.Clone(nums), 2, 100, 200)
	fmt.Println("slices.Insert:", inserted)

	deleted := slices.Delete(slices.Clone(nums), 1, 3)
	fmt.Println("slices.Delete:", deleted)

	stack := []int{1, 2, 3}
	top := stack[len(stack)-1]
	stack = stack[:len(stack)-1]
	fmt.Println("pop:", top, "remaining:", stack)

	grid := make([][]int, 2)
	for i := range grid {
		grid[i] = make([]int, 3)
		for j := range grid[i] {
			grid[i][j] = i*3 + j
		}
	}
	fmt.Println("2d slice:", grid)

	sum := 0
	for v := range slices.Values([]int{10, 20, 30}) {
		sum += v
	}
	fmt.Println("slices.Values sum:", sum)
}
