package complete

import (
	"os/exec"
	"strings"
	"testing"
)

func build(t *testing.T) *Index {
	t.Helper()
	out, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		t.Skip("go toolchain not found")
	}
	idx, err := Build(strings.TrimSpace(string(out)))
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return idx
}

func labels(items []Item) map[string]Item {
	m := map[string]Item{}
	for _, it := range items {
		m[it.Label] = it
	}
	return m
}

func TestQuery(t *testing.T) {
	idx := build(t)
	tests := []struct {
		name     string
		prefix   string
		want     string
		wantKind Kind
	}{
		{"package members come from the real stdlib source", "fmt.Pri", "Println", Func},
		{"concurrency types are offered for sync", "sync.Wait", "WaitGroup", TypeKind},
		{"context helpers are offered", "context.WithCan", "WithCancel", Func},
		{"keywords complete without a package", "sel", "select", Keyword},
		{"builtins complete without a package", "mak", "make", Builtin},
		{"package names complete so users can reach members", "str", "strings", Package},
		{"snippets speed up common goroutine code", "gof", "gofunc", Snippet},
		{"nested package uses its last path element", "atomic.Add", "AddInt64", Func},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := labels(idx.Query(tt.prefix, 50))[tt.want]
			if !ok {
				t.Fatalf("Query(%q) missing %q", tt.prefix, tt.want)
			}
			if got.Kind != tt.wantKind {
				t.Errorf("kind = %s, want %s", got.Kind, tt.wantKind)
			}
		})
	}
}

func TestQueryHidesUnexportedNamesSoUsersOnlySeeUsableAPI(t *testing.T) {
	idx := build(t)
	for _, it := range idx.Query("fmt.", 1000) {
		if it.Label[0] >= 'a' && it.Label[0] <= 'z' {
			t.Errorf("unexported %q offered", it.Label)
		}
	}
}

func TestQueryRespectsLimit(t *testing.T) {
	idx := build(t)
	if got := len(idx.Query("", 5)); got != 5 {
		t.Errorf("len = %d, want 5", got)
	}
}

func TestQueryUnknownPackageReturnsEmpty(t *testing.T) {
	idx := build(t)
	if got := idx.Query("nopkg.X", 10); len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}

func TestSignatureShowsParams(t *testing.T) {
	idx := build(t)
	got := labels(idx.Query("strings.Repl", 10))["ReplaceAll"]
	if got.Detail != "func(string, string, string)" {
		t.Errorf("detail = %q", got.Detail)
	}
}
