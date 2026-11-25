package markdown

import (
	"html"
	"regexp"
	"strings"
)

var (
	boldRe   = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	italicRe = regexp.MustCompile(`(^|[^*])\*([^*\s][^*]*)\*`)
	linkRe   = regexp.MustCompile(`\[([^\]]+)\]\(([^)\s]+)\)`)
	olRe     = regexp.MustCompile(`^\d+\.\s+`)
	headRe   = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
	slugRe   = regexp.MustCompile(`[^a-z0-9]+`)
)

type renderer struct {
	lines []string
	pos   int
	out   strings.Builder
}

func Render(src string) string {
	r := &renderer{lines: strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")}
	for r.pos < len(r.lines) {
		r.block()
	}
	return r.out.String()
}

func (r *renderer) block() {
	line := r.lines[r.pos]
	trimmed := strings.TrimSpace(line)
	switch {
	case trimmed == "":
		r.pos++
	case strings.HasPrefix(trimmed, "```"):
		r.code(strings.TrimSpace(strings.TrimPrefix(trimmed, "```")))
	case headRe.MatchString(trimmed):
		m := headRe.FindStringSubmatch(trimmed)
		lvl := string(rune('0' + len(m[1])))
		r.out.WriteString("<h" + lvl + " id=\"" + Slug(m[2]) + "\">" + Inline(m[2]) + "</h" + lvl + ">\n")
		r.pos++
	case trimmed == "---":
		r.out.WriteString("<hr>\n")
		r.pos++
	case strings.HasPrefix(trimmed, ">"):
		r.quote()
	case strings.HasPrefix(trimmed, "|"):
		r.table()
	case strings.HasPrefix(trimmed, "- "), strings.HasPrefix(trimmed, "* "):
		r.list("ul", func(s string) (string, bool) {
			if strings.HasPrefix(s, "- ") || strings.HasPrefix(s, "* ") {
				return s[2:], true
			}
			return "", false
		})
	case olRe.MatchString(trimmed):
		r.list("ol", func(s string) (string, bool) {
			if loc := olRe.FindStringIndex(s); loc != nil {
				return s[loc[1]:], true
			}
			return "", false
		})
	default:
		r.paragraph()
	}
}

func (r *renderer) code(lang string) {
	r.pos++
	var body []string
	for r.pos < len(r.lines) && !strings.HasPrefix(strings.TrimSpace(r.lines[r.pos]), "```") {
		body = append(body, r.lines[r.pos])
		r.pos++
	}
	r.pos++
	if lang == "" {
		lang = "text"
	}
	r.out.WriteString("<pre class=\"code\" data-lang=\"" + html.EscapeString(lang) + "\"><code>" + html.EscapeString(strings.Join(body, "\n")) + "</code></pre>\n")
}

func (r *renderer) quote() {
	var body []string
	for r.pos < len(r.lines) && strings.HasPrefix(strings.TrimSpace(r.lines[r.pos]), ">") {
		body = append(body, strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(r.lines[r.pos]), ">")))
		r.pos++
	}
	class := "note"
	if len(body) > 0 {
		for _, k := range []string{"TIP", "WARNING", "NOTE", "INFO"} {
			tag := "[!" + k + "]"
			if strings.HasPrefix(body[0], tag) {
				class = strings.ToLower(k)
				body[0] = strings.TrimSpace(strings.TrimPrefix(body[0], tag))
			}
		}
	}
	r.out.WriteString("<blockquote class=\"callout " + class + "\"><p>" + Inline(strings.TrimSpace(strings.Join(body, " "))) + "</p></blockquote>\n")
}

func (r *renderer) table() {
	var rows [][]string
	for r.pos < len(r.lines) && strings.HasPrefix(strings.TrimSpace(r.lines[r.pos]), "|") {
		rows = append(rows, cells(r.lines[r.pos]))
		r.pos++
	}
	r.out.WriteString("<table>\n")
	for i, row := range rows {
		if i == 1 && isSeparator(row) {
			continue
		}
		tag := "td"
		if i == 0 {
			tag = "th"
		}
		r.out.WriteString("<tr>")
		for _, c := range row {
			r.out.WriteString("<" + tag + ">" + Inline(c) + "</" + tag + ">")
		}
		r.out.WriteString("</tr>\n")
	}
	r.out.WriteString("</table>\n")
}

func (r *renderer) list(tag string, item func(string) (string, bool)) {
	r.out.WriteString("<" + tag + ">\n")
	for r.pos < len(r.lines) {
		text, ok := item(strings.TrimSpace(r.lines[r.pos]))
		if !ok {
			break
		}
		r.out.WriteString("<li>" + Inline(text) + "</li>\n")
		r.pos++
	}
	r.out.WriteString("</" + tag + ">\n")
}

func (r *renderer) paragraph() {
	var body []string
	for r.pos < len(r.lines) {
		t := strings.TrimSpace(r.lines[r.pos])
		if t == "" || strings.HasPrefix(t, "#") || strings.HasPrefix(t, "```") || strings.HasPrefix(t, ">") || strings.HasPrefix(t, "|") || strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ") || olRe.MatchString(t) {
			break
		}
		body = append(body, t)
		r.pos++
	}
	r.out.WriteString("<p>" + Inline(strings.Join(body, " ")) + "</p>\n")
}

func Inline(s string) string {
	parts := strings.Split(s, "`")
	var b strings.Builder
	for i, p := range parts {
		if i%2 == 1 && i < len(parts)-1 {
			b.WriteString("<code>" + html.EscapeString(p) + "</code>")
			continue
		}
		if i%2 == 1 {
			b.WriteString("`")
		}
		e := html.EscapeString(p)
		e = linkRe.ReplaceAllString(e, `<a href="$2" target="_blank" rel="noopener">$1</a>`)
		e = boldRe.ReplaceAllString(e, "<strong>$1</strong>")
		e = italicRe.ReplaceAllString(e, "$1<em>$2</em>")
		b.WriteString(e)
	}
	return b.String()
}

func Slug(s string) string {
	return strings.Trim(slugRe.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

func cells(line string) []string {
	t := strings.Trim(strings.TrimSpace(line), "|")
	parts := strings.Split(t, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func isSeparator(row []string) bool {
	for _, c := range row {
		if strings.Trim(c, "-: ") != "" {
			return false
		}
	}
	return true
}
