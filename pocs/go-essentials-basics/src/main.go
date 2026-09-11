package main

import (
	"fmt"
	"sort"
)

type Animal struct {
	Name string
	Legs int
}

func (a Animal) Describe() string {
	return fmt.Sprintf("%s has %d legs", a.Name, a.Legs)
}

type Dog struct {
	Animal
}

func NewDog(name string) Dog {
	return Dog{Animal{Name: name, Legs: 4}}
}

func (d Dog) Sound() string {
	return "woof"
}

func forLoop() int {
	total := 0
	for i := 1; i <= 5; i++ {
		total += i
	}
	return total
}

func whileLoop() int {
	n, steps := 10, 0
	for n > 1 {
		n /= 2
		steps++
	}
	return steps
}

func loopOverCollection() []string {
	fruits := []string{"apple", "banana", "cherry"}
	out := make([]string, 0, len(fruits))
	for i, fruit := range fruits {
		out = append(out, fmt.Sprintf("%d:%s", i, fruit))
	}
	return out
}

func mapsAndSwitch(n int) string {
	grades := map[int]string{1: "one", 2: "two", 3: "three"}
	switch {
	case n <= 0:
		return "non-positive"
	case n <= 3:
		return grades[n]
	default:
		return "many"
	}
}

type person struct {
	name string
	age  int
}

func sorting() ([]person, []person) {
	people := []person{{"alice", 30}, {"bob", 25}, {"carol", 35}}
	byAge := append([]person(nil), people...)
	sort.Slice(byAge, func(i, j int) bool { return byAge[i].age < byAge[j].age })
	byNameDesc := append([]person(nil), people...)
	sort.Slice(byNameDesc, func(i, j int) bool { return byNameDesc[i].name > byNameDesc[j].name })
	return byAge, byNameDesc
}

func main() {
	fmt.Println("forLoop sum 1..5:", forLoop())
	fmt.Println("whileLoop halving steps:", whileLoop())
	fmt.Println("loopOverCollection:", loopOverCollection())
	fmt.Println("mapsAndSwitch(2):", mapsAndSwitch(2))
	fmt.Println("mapsAndSwitch(9):", mapsAndSwitch(9))

	byAge, byNameDesc := sorting()
	fmt.Println("sorted by age:", byAge)
	fmt.Println("sorted by name desc:", byNameDesc)

	dog := NewDog("Rex")
	fmt.Println("struct describe:", dog.Describe())
	fmt.Println("struct method:", dog.Sound())
	fmt.Println("embedded field:", dog.Name)
}
