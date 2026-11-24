package match

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const Reset = "\x1b[0m"

var palette = []string{
	"\x1b[1;31m",
	"\x1b[1;33m",
	"\x1b[1;32m",
	"\x1b[1;36m",
	"\x1b[1;35m",
	"\x1b[1;34m",
}

type Pattern struct {
	Expr  string
	Color string
	re    *regexp.Regexp
}

type Result struct {
	Hits []int
	Text string
}

type Matcher struct {
	patterns []Pattern
	color    bool
}

type span struct {
	start, end, pattern int
}

func New(exprs []string, color bool) (*Matcher, error) {
	m := &Matcher{color: color}
	for i, expr := range exprs {
		re, err := regexp.Compile(expr)
		if err != nil {
			return nil, fmt.Errorf("pattern %q: %w", expr, err)
		}
		m.patterns = append(m.patterns, Pattern{Expr: expr, Color: palette[i%len(palette)], re: re})
	}
	return m, nil
}

func (m *Matcher) Patterns() []Pattern {
	return m.patterns
}

func (m *Matcher) Paint(text string, pattern int) string {
	if !m.color {
		return text
	}
	return m.patterns[pattern].Color + text + Reset
}

func (m *Matcher) Match(line string) Result {
	var spans []span
	var hits []int
	for i, p := range m.patterns {
		locs := p.re.FindAllStringIndex(line, -1)
		if len(locs) == 0 {
			continue
		}
		hits = append(hits, i)
		for _, loc := range locs {
			if loc[1] > loc[0] {
				spans = append(spans, span{start: loc[0], end: loc[1], pattern: i})
			}
		}
	}
	if !m.color || len(spans) == 0 {
		return Result{Hits: hits, Text: line}
	}
	return Result{Hits: hits, Text: m.highlight(line, spans)}
}

func (m *Matcher) highlight(line string, spans []span) string {
	sort.SliceStable(spans, func(a, b int) bool { return spans[a].start < spans[b].start })
	var sb strings.Builder
	pos := 0
	for _, s := range spans {
		if s.start < pos {
			continue
		}
		sb.WriteString(line[pos:s.start])
		sb.WriteString(m.Paint(line[s.start:s.end], s.pattern))
		pos = s.end
	}
	sb.WriteString(line[pos:])
	return sb.String()
}
