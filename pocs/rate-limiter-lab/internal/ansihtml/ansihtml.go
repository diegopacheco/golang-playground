package ansihtml

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	CellWidth  = 9.1
	CellHeight = 19
	Padding    = 24
)

type Theme struct {
	Background string
	Foreground string
}

var Dark = Theme{Background: "#14161f", Foreground: "#e4e6ef"}

var basic = [16]string{
	"#1d1f28", "#f37f97", "#5adecd", "#f2a272", "#8897f4", "#c574dd", "#79e6f3", "#c0c3d0",
	"#5c6077", "#ff4971", "#18e3c8", "#ff8037", "#556fff", "#b043d1", "#3fdcee", "#fdfdfe",
}

type style struct {
	fg, bg                         string
	bold, faint, italic, underline bool
	reverse                        bool
}

func (s style) css(theme Theme) string {
	fg, bg := s.fg, s.bg
	if s.reverse {
		fg, bg = bg, fg
		if fg == "" {
			fg = theme.Background
		}
		if bg == "" {
			bg = theme.Foreground
		}
	}
	var parts []string
	if fg != "" {
		parts = append(parts, "color:"+fg)
	}
	if bg != "" {
		parts = append(parts, "background:"+bg)
	}
	if s.bold {
		parts = append(parts, "font-weight:700")
	}
	if s.faint {
		parts = append(parts, "opacity:.6")
	}
	if s.italic {
		parts = append(parts, "font-style:italic")
	}
	if s.underline {
		parts = append(parts, "text-decoration:underline")
	}
	return strings.Join(parts, ";")
}

type Screen struct {
	Body string
	Cols int
	Rows int
}

func Convert(content string, theme Theme) Screen {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	var b strings.Builder
	cols := 0
	var st style
	for i, line := range lines {
		if i > 0 {
			b.WriteByte('\n')
		}
		var width int
		st, width = convertLine(&b, line, st, theme)
		cols = max(cols, width)
	}
	return Screen{Body: b.String(), Cols: cols, Rows: len(lines)}
}

func convertLine(b *strings.Builder, line string, st style, theme Theme) (style, int) {
	var run strings.Builder
	width := 0
	flush := func() {
		if run.Len() == 0 {
			return
		}
		text := html.EscapeString(run.String())
		if css := st.css(theme); css != "" {
			fmt.Fprintf(b, `<span style="%s">%s</span>`, css, text)
		} else {
			b.WriteString(text)
		}
		run.Reset()
	}
	for i := 0; i < len(line); {
		if line[i] == 0x1b && i+1 < len(line) {
			flush()
			next, params, final := escape(line, i)
			if final == 'm' {
				st = applySGR(st, params)
			}
			i = next
			continue
		}
		r, size := utf8.DecodeRuneInString(line[i:])
		i += size
		if r == '\r' {
			continue
		}
		run.WriteRune(r)
		width++
	}
	flush()
	return st, width
}

func escape(s string, i int) (int, string, byte) {
	switch s[i+1] {
	case '[':
		for j := i + 2; j < len(s); j++ {
			if s[j] >= 0x40 && s[j] <= 0x7e {
				return j + 1, s[i+2 : j], s[j]
			}
		}
		return len(s), "", 0
	case ']':
		for j := i + 2; j < len(s); j++ {
			if s[j] == 0x07 {
				return j + 1, "", 0
			}
			if s[j] == 0x1b && j+1 < len(s) && s[j+1] == '\\' {
				return j + 2, "", 0
			}
		}
		return len(s), "", 0
	}
	return i + 2, "", 0
}

func applySGR(st style, params string) style {
	if params == "" {
		return style{}
	}
	var codes []int
	for field := range strings.FieldsFuncSeq(params, func(r rune) bool { return r == ';' || r == ':' }) {
		n, err := strconv.Atoi(field)
		if err == nil {
			codes = append(codes, n)
		}
	}
	for i := 0; i < len(codes); i++ {
		c := codes[i]
		switch {
		case c == 0:
			st = style{}
		case c == 1:
			st.bold = true
		case c == 2:
			st.faint = true
		case c == 3:
			st.italic = true
		case c == 4:
			st.underline = true
		case c == 7:
			st.reverse = true
		case c == 22:
			st.bold, st.faint = false, false
		case c == 23:
			st.italic = false
		case c == 24:
			st.underline = false
		case c == 27:
			st.reverse = false
		case c >= 30 && c <= 37:
			st.fg = basic[c-30]
		case c >= 90 && c <= 97:
			st.fg = basic[c-90+8]
		case c >= 40 && c <= 47:
			st.bg = basic[c-40]
		case c >= 100 && c <= 107:
			st.bg = basic[c-100+8]
		case c == 39:
			st.fg = ""
		case c == 49:
			st.bg = ""
		case c == 38 || c == 48:
			color, used := extended(codes[i+1:])
			if c == 38 {
				st.fg = color
			} else {
				st.bg = color
			}
			i += used
		}
	}
	return st
}

func extended(codes []int) (string, int) {
	if len(codes) >= 4 && codes[0] == 2 {
		return fmt.Sprintf("#%02x%02x%02x", codes[1], codes[2], codes[3]), 4
	}
	if len(codes) >= 2 && codes[0] == 5 {
		return palette256(codes[1]), 2
	}
	return "", len(codes)
}

func palette256(n int) string {
	switch {
	case n < 16:
		return basic[n]
	case n < 232:
		n -= 16
		level := func(v int) int {
			if v == 0 {
				return 0
			}
			return 55 + v*40
		}
		return fmt.Sprintf("#%02x%02x%02x", level(n/36), level(n/6%6), level(n%6))
	case n < 256:
		g := 8 + (n-232)*10
		return fmt.Sprintf("#%02x%02x%02x", g, g, g)
	}
	return ""
}

func Page(title string, content string, theme Theme) string {
	sc := Convert(content, theme)
	w := int(float64(sc.Cols)*CellWidth) + 2*Padding
	h := sc.Rows*CellHeight + 2*Padding
	return fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>%s</title>
<style>html,body{margin:0;background:%s}pre{margin:0;padding:%dpx;color:%s;font:15px/%dpx Menlo,'SF Mono',Monaco,monospace;white-space:pre}</style>
</head><body data-size="%d,%d"><pre>%s</pre></body></html>
`, html.EscapeString(title), theme.Background, Padding, theme.Foreground, CellHeight, w, h, sc.Body)
}
