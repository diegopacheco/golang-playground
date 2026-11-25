package markdown

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{"headings get anchors so the table of contents can link them", "## Buffered Channels", `<h2 id="buffered-channels">Buffered Channels</h2>`},
		{"go code blocks keep the language for the highlighter", "```go\nch := make(chan int)\n```", `<pre class="code" data-lang="go"><code>ch := make(chan int)</code></pre>`},
		{"code is escaped so arrows and generics display as text", "```go\nv := <-ch\n```", "v := &lt;-ch"},
		{"inline code is escaped", "send with `ch <- v` now", "<code>ch &lt;- v</code>"},
		{"bold text", "a **must** read", "<strong>must</strong>"},
		{"italic text", "an *easy* one", "<em>easy</em>"},
		{"links open in a new window", "[Go](https://go.dev)", `<a href="https://go.dev" target="_blank" rel="noopener">Go</a>`},
		{"unordered lists", "- one\n- two", "<ul>\n<li>one</li>\n<li>two</li>\n</ul>"},
		{"ordered lists", "1. one\n2. two", "<ol>\n<li>one</li>\n<li>two</li>\n</ol>"},
		{"tables skip the separator row", "| a | b |\n|---|---|\n| 1 | 2 |", "<table>\n<tr><th>a</th><th>b</th></tr>\n<tr><td>1</td><td>2</td></tr>\n</table>"},
		{"callouts get their kind as class", "> [!WARNING] closing twice panics", `<blockquote class="callout warning"><p>closing twice panics</p></blockquote>`},
		{"plain quotes become notes", "> remember this", `<blockquote class="callout note"><p>remember this</p></blockquote>`},
		{"paragraph lines join", "line one\nline two", "<p>line one line two</p>"},
		{"raw html in text is escaped so docs cannot inject script", "<script>x</script>", "&lt;script&gt;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(tt.src)
			if !strings.Contains(got, tt.want) {
				t.Errorf("Render(%q) = %q, want it to contain %q", tt.src, got, tt.want)
			}
		})
	}
}

func TestRenderDoesNotFormatInsideCode(t *testing.T) {
	got := Render("```go\nx := a**b**\n```")
	if strings.Contains(got, "<strong>") {
		t.Errorf("formatted inside code block: %q", got)
	}
}
