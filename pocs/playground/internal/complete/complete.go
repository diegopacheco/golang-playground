package complete

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Kind string

const (
	Keyword  Kind = "keyword"
	Builtin  Kind = "builtin"
	Package  Kind = "package"
	Func     Kind = "func"
	TypeKind Kind = "type"
	Var      Kind = "var"
	Const    Kind = "const"
	Snippet  Kind = "snippet"
)

type Item struct {
	Label  string `json:"label"`
	Kind   Kind   `json:"kind"`
	Detail string `json:"detail"`
	Insert string `json:"insert,omitempty"`
}

var Packages = []string{
	"fmt", "strings", "strconv", "errors", "time", "sync", "sync/atomic", "context",
	"os", "io", "bufio", "bytes", "math", "math/rand", "sort", "slices", "maps",
	"unicode", "encoding/json", "net/http", "runtime", "log", "log/slog", "iter",
}

var keywords = []string{
	"break", "case", "chan", "const", "continue", "default", "defer", "else",
	"fallthrough", "for", "func", "go", "goto", "if", "import", "interface",
	"map", "package", "range", "return", "select", "struct", "switch", "type", "var",
}

var builtins = []string{
	"append", "cap", "clear", "close", "complex", "copy", "delete", "imag", "len",
	"make", "max", "min", "new", "panic", "print", "println", "real", "recover",
	"bool", "byte", "error", "float32", "float64", "int", "int8", "int16", "int32",
	"int64", "rune", "string", "uint", "uint8", "uint16", "uint32", "uint64",
	"uintptr", "any", "comparable", "true", "false", "nil", "iota",
}

var snippets = []Item{
	{Label: "gofunc", Kind: Snippet, Detail: "start a goroutine", Insert: "go func() {\n\t\n}()"},
	{Label: "forrange", Kind: Snippet, Detail: "range loop", Insert: "for i, v := range items {\n\t\n}"},
	{Label: "selectch", Kind: Snippet, Detail: "select on channels", Insert: "select {\ncase v := <-ch:\n\t_ = v\ncase <-ctx.Done():\n\treturn\n}"},
	{Label: "iferr", Kind: Snippet, Detail: "error check", Insert: "if err != nil {\n\treturn err\n}"},
	{Label: "mainpkg", Kind: Snippet, Detail: "main package skeleton", Insert: "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}"},
}

type Index struct {
	global  []Item
	members map[string][]Item
}

func Build(goroot string) (*Index, error) {
	idx := &Index{members: map[string][]Item{}}
	for _, k := range keywords {
		idx.global = append(idx.global, Item{Label: k, Kind: Keyword, Detail: "keyword"})
	}
	for _, b := range builtins {
		idx.global = append(idx.global, Item{Label: b, Kind: Builtin, Detail: "builtin"})
	}
	idx.global = append(idx.global, snippets...)
	for _, p := range Packages {
		name := filepath.Base(p)
		idx.global = append(idx.global, Item{Label: name, Kind: Package, Detail: "package " + p})
		items, err := exported(filepath.Join(goroot, "src", filepath.FromSlash(p)))
		if err != nil {
			return nil, err
		}
		idx.members[name] = items
	}
	sortItems(idx.global)
	return idx, nil
}

func (idx *Index) Query(prefix string, limit int) []Item {
	pool, word := idx.global, prefix
	if pkg, rest, ok := strings.Cut(prefix, "."); ok {
		pool, word = idx.members[pkg], rest
	}
	out := []Item{}
	for _, it := range pool {
		if strings.HasPrefix(it.Label, word) {
			out = append(out, it)
			if len(out) == limit {
				break
			}
		}
	}
	return out
}

func (idx *Index) Size() int {
	n := len(idx.global)
	for _, m := range idx.members {
		n += len(m)
	}
	return n
}

func exported(dir string) ([]Item, error) {
	fset := token.NewFileSet()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var items []Item
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, n), nil, parser.SkipObjectResolution)
		if err != nil {
			continue
		}
		if f.Name.Name == "main" || strings.HasSuffix(f.Name.Name, "_test") {
			continue
		}
		for _, it := range declItems(f) {
			if !seen[it.Label] {
				seen[it.Label] = true
				items = append(items, it)
			}
		}
	}
	sortItems(items)
	return items, nil
}

func declItems(f *ast.File) []Item {
	var items []Item
	for _, d := range f.Decls {
		switch decl := d.(type) {
		case *ast.FuncDecl:
			if decl.Recv == nil && decl.Name.IsExported() {
				items = append(items, Item{Label: decl.Name.Name, Kind: Func, Detail: signature(decl)})
			}
		case *ast.GenDecl:
			items = append(items, genItems(decl)...)
		}
	}
	return items
}

func genItems(decl *ast.GenDecl) []Item {
	var items []Item
	for _, s := range decl.Specs {
		switch spec := s.(type) {
		case *ast.TypeSpec:
			if spec.Name.IsExported() {
				items = append(items, Item{Label: spec.Name.Name, Kind: TypeKind, Detail: "type"})
			}
		case *ast.ValueSpec:
			kind := Var
			if decl.Tok == token.CONST {
				kind = Const
			}
			for _, n := range spec.Names {
				if n.IsExported() {
					items = append(items, Item{Label: n.Name, Kind: kind, Detail: string(kind)})
				}
			}
		}
	}
	return items
}

func signature(fn *ast.FuncDecl) string {
	var params []string
	for _, p := range fn.Type.Params.List {
		for range max(1, len(p.Names)) {
			params = append(params, exprString(p.Type))
		}
	}
	return "func(" + strings.Join(params, ", ") + ")"
}

func exprString(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprString(t.X)
	case *ast.SelectorExpr:
		return exprString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprString(t.Elt)
	case *ast.Ellipsis:
		return "..." + exprString(t.Elt)
	case *ast.MapType:
		return "map[" + exprString(t.Key) + "]" + exprString(t.Value)
	case *ast.InterfaceType:
		return "any"
	case *ast.ChanType:
		return "chan " + exprString(t.Value)
	case *ast.FuncType:
		return "func"
	case *ast.IndexExpr:
		return exprString(t.X) + "[" + exprString(t.Index) + "]"
	default:
		return "T"
	}
}

func sortItems(items []Item) {
	sort.SliceStable(items, func(i, j int) bool { return items[i].Label < items[j].Label })
}
