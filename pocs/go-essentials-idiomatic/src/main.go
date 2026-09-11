package main

import (
	"fmt"
	"strings"
)

type Server struct {
	host    string
	port    int
	tls     bool
	timeout int
}

type Option func(*Server)

func WithHost(h string) Option    { return func(s *Server) { s.host = h } }
func WithPort(p int) Option       { return func(s *Server) { s.port = p } }
func WithTLS() Option             { return func(s *Server) { s.tls = true } }
func WithTimeout(t int) Option    { return func(s *Server) { s.timeout = t } }

func NewServer(opts ...Option) *Server {
	s := &Server{host: "localhost", port: 8080, timeout: 30}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Server) String() string {
	scheme := "http"
	if s.tls {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s:%d timeout=%ds", scheme, s.host, s.port, s.timeout)
}

func commaOk() string {
	m := map[string]int{"a": 1}
	if v, ok := m["a"]; ok {
		return fmt.Sprintf("found %d", v)
	}
	return "missing"
}

func deferCleanup() string {
	var steps []string
	release := func(name string) func() {
		steps = append(steps, "acquire "+name)
		return func() { steps = append(steps, "release "+name) }
	}
	func() {
		defer release("db")()
		defer release("lock")()
		steps = append(steps, "work")
	}()
	return strings.Join(steps, " -> ")
}

func zeroValues() string {
	var i int
	var s string
	var b bool
	var sl []int
	return fmt.Sprintf("int=%d str=%q bool=%v slice=%v nil=%v", i, s, b, sl, sl == nil)
}

func main() {
	fmt.Println("default:", NewServer())
	fmt.Println("configured:", NewServer(WithHost("api.local"), WithPort(443), WithTLS(), WithTimeout(10)))
	fmt.Println("comma-ok:", commaOk())
	fmt.Println("defer cleanup:", deferCleanup())
	fmt.Println("zero values:", zeroValues())
}
