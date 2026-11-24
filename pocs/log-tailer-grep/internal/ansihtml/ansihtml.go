package ansihtml

import (
	"bufio"
	"fmt"
	"html"
	"io"
	"strconv"
	"strings"
)

var colors = map[int]string{
	30: "#4b5263", 31: "#ff6b6b", 32: "#7ee787", 33: "#ffd166",
	34: "#79c0ff", 35: "#e599f7", 36: "#66e3e3", 37: "#e6e6e6",
	90: "#8b949e", 91: "#ff8787", 92: "#a5f3a5", 93: "#ffe08a",
	94: "#a5d6ff", 95: "#f0b3ff", 96: "#99f0f0", 97: "#ffffff",
}

type style struct {
	bold  bool
	color string
}

func (s style) css() string {
	var parts []string
	if s.color != "" {
		parts = append(parts, "color:"+s.color)
	}
	if s.bold {
		parts = append(parts, "font-weight:700")
	}
	return strings.Join(parts, ";")
}

func (s style) apply(codes string) style {
	for _, raw := range strings.Split(codes, ";") {
		code, err := strconv.Atoi(raw)
		if raw == "" || err != nil {
			code = 0
		}
		switch {
		case code == 0:
			s = style{}
		case code == 1:
			s.bold = true
		case colors[code] != "":
			s.color = colors[code]
		}
	}
	return s
}

func Convert(r io.Reader, w io.Writer, title string) error {
	var body strings.Builder
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		body.WriteString(line(sc.Text()))
		body.WriteString("\n")
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("scan ansi input: %w", err)
	}
	_, err := fmt.Fprintf(w, page, html.EscapeString(title), body.String())
	if err != nil {
		return fmt.Errorf("write html: %w", err)
	}
	return nil
}

func line(text string) string {
	var sb strings.Builder
	cur := style{}
	open := false
	for len(text) > 0 {
		i := strings.Index(text, "\x1b[")
		if i < 0 {
			sb.WriteString(html.EscapeString(text))
			break
		}
		sb.WriteString(html.EscapeString(text[:i]))
		rest := text[i+2:]
		end := strings.IndexByte(rest, 'm')
		if end < 0 {
			break
		}
		cur = cur.apply(rest[:end])
		text = rest[end+1:]
		if open {
			sb.WriteString("</span>")
			open = false
		}
		if css := cur.css(); css != "" {
			fmt.Fprintf(&sb, `<span style="%s">`, css)
			open = true
		}
	}
	if open {
		sb.WriteString("</span>")
	}
	return sb.String()
}

const page = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>%s</title>
<style>body{margin:0;background:#0d1117}pre{margin:0;padding:18px;color:#c9d1d9;font:14px/1.45 Menlo,Monaco,monospace;white-space:pre}</style>
</head><body><pre>%s</pre></body></html>
`
