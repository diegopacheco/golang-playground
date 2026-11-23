package ansihtml

import (
	"strings"
	"testing"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     string
		wantCols int
		wantRows int
	}{
		{"plain text passes through so unstyled output is not lost", "hello", "hello", 5, 1},
		{"truecolor foreground becomes a css color so lipgloss colors survive the screenshot", "\x1b[38;2;255;95;210mhi\x1b[m", `<span style="color:#ff5fd2">hi</span>`, 2, 1},
		{"colon separated truecolor is also understood", "\x1b[48:2::0:0:255mx\x1b[0m", `<span style="background:#0000ff">x</span>`, 1, 1},
		{"html characters are escaped so captured text cannot break the page", "<a&b>", "&lt;a&amp;b&gt;", 5, 1},
		{"escape codes take no width so the viewport matches the visible text", "\x1b[1mab\x1b[22m\ncdef", "<span style=\"font-weight:700\">ab</span>\ncdef", 4, 2},
		{"cursor and hyperlink sequences are dropped instead of printed as garbage", "\x1b[?25la\x1b]8;;http://x\x07b", "ab", 2, 1},
		{"256 color cube indexes map to their rgb value", "\x1b[38;5;196mr", `<span style="color:#ff0000">r</span>`, 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Convert(tt.input, Dark)
			if got.Body != tt.want {
				t.Fatalf("body = %q, want %q", got.Body, tt.want)
			}
			if got.Cols != tt.wantCols || got.Rows != tt.wantRows {
				t.Fatalf("size = %dx%d, want %dx%d", got.Cols, got.Rows, tt.wantCols, tt.wantRows)
			}
		})
	}
}

func TestPageCarriesItsPixelSizeForTheScreenshotViewport(t *testing.T) {
	page := Page("t", "abcd\nef", Dark)
	want := `data-size="84,86"`
	if !strings.Contains(page, want) {
		t.Fatalf("page is missing %s:\n%s", want, page)
	}
}
