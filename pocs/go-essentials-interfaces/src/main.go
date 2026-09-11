package main

import "fmt"

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64      { return 3.14159 * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * 3.14159 * c.Radius }

type Stringer interface {
	String() string
}

func (c Circle) String() string { return fmt.Sprintf("Circle(r=%.0f)", c.Radius) }

func describe(s Shape) string {
	return fmt.Sprintf("area=%.2f perimeter=%.2f", s.Area(), s.Perimeter())
}

func typeSwitch(v any) string {
	switch x := v.(type) {
	case int:
		return fmt.Sprintf("int:%d", x)
	case string:
		return fmt.Sprintf("string:%s", x)
	case Shape:
		return "shape"
	default:
		return "unknown"
	}
}

type ReadCloser interface {
	Reader
	Closer
}

type Reader interface{ Read() string }
type Closer interface{ Close() string }

type file struct{ name string }

func (f file) Read() string  { return "data from " + f.name }
func (f file) Close() string { return "closed " + f.name }

func main() {
	shapes := []Shape{Rectangle{3, 4}, Circle{5}}
	for _, s := range shapes {
		fmt.Println("describe:", describe(s))
	}

	c := Circle{2}
	if str, ok := any(c).(Stringer); ok {
		fmt.Println("type assertion:", str.String())
	}

	fmt.Println("typeSwitch int:", typeSwitch(42))
	fmt.Println("typeSwitch string:", typeSwitch("hi"))
	fmt.Println("typeSwitch shape:", typeSwitch(Rectangle{1, 1}))

	var rc ReadCloser = file{"log.txt"}
	fmt.Println("embedded interface:", rc.Read(), "|", rc.Close())
}
