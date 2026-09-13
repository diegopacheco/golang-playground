package theme

import "github.com/gdamore/tcell/v2"

const (
	Bg          = "#0b0f1a"
	Panel       = "#11172a"
	Border      = "#2a3355"
	Cyan        = "#00e5ff"
	Magenta     = "#ff4fd8"
	Lime        = "#a6ff4d"
	Orange      = "#ffb86c"
	Purple      = "#bd93f9"
	Red         = "#ff5c7a"
	Yellow      = "#ffe66d"
	Dim         = "#6b7394"
	Text        = "#d7def0"
	LightSquare = "#8e9cc4"
	DarkSquare  = "#4b5784"
	WhitePiece  = "#fffaf0"
	BlackPiece  = "#0b0f1a"
	BlackGlyph  = "#9aa3c7"
)

func Color(hex string) tcell.Color {
	return tcell.GetColor(hex)
}

func Mix(a, b string, t float64) tcell.Color {
	ar, ag, ab := Color(a).RGB()
	br, bg, bb := Color(b).RGB()
	mix := func(x, y int32) int32 { return x + int32(float64(y-x)*t) }
	return tcell.NewRGBColor(mix(ar, br), mix(ag, bg), mix(ab, bb))
}
