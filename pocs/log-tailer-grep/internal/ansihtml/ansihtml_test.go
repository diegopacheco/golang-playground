package ansihtml

import (
	"strings"
	"testing"
)

func TestConvert(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "bold red becomes a styled span so the screenshot matches the terminal",
			input: "x \x1b[1;31mERROR\x1b[0m y",
			want:  `x <span style="color:#ff6b6b;font-weight:700">ERROR</span> y`,
		},
		{
			name:  "html characters are escaped so log content cannot break the page",
			input: "<b>&</b>",
			want:  "&lt;b&gt;&amp;&lt;/b&gt;",
		},
		{
			name:  "an unclosed color is closed at end of line so spans never leak",
			input: "\x1b[90mapp.log |",
			want:  `<span style="color:#8b949e">app.log |</span>`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out strings.Builder
			if err := Convert(strings.NewReader(tc.input), &out, "t"); err != nil {
				t.Fatalf("convert: %v", err)
			}
			if !strings.Contains(out.String(), tc.want) {
				t.Fatalf("missing %q in %q", tc.want, out.String())
			}
		})
	}
}
