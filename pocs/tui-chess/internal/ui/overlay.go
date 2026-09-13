package ui

import (
	"strings"

	"github.com/diegopacheco/tui-chess/internal/chess"
	"github.com/diegopacheco/tui-chess/internal/theme"
)

var promotionOrder = []chess.Kind{chess.Queen, chess.Rook, chess.Bishop, chess.Knight}

var helpLines = []hint{
	{"←↑↓→  hjkl", "move the cursor"},
	{"enter space", "select a piece and move it"},
	{"mouse click", "select a piece and move it"},
	{"esc", "clear selection or hide result"},
	{"u", "undo your last move"},
	{"f", "flip the board"},
	{"n", "new game"},
	{"m", "back to the menu"},
	{"?", "this help"},
	{"q  ctrl-c", "quit"},
}

func (a *App) drawHelp(w, h int) {
	s := a.screen
	bg := theme.Color(theme.Panel)
	width, height := 50, len(helpLines)+6
	x, y := (w-width)/2, (h-height)/2
	titledFrame(s, x, y, width, height, theme.Color(theme.Cyan), bg, "KEYS", theme.Color(theme.Cyan))
	for i, line := range helpLines {
		put(s, x+3, y+2+i, line.key, style(theme.Color(theme.Lime), bg).Bold(true))
		put(s, x+17, y+2+i, line.text, style(theme.Color(theme.Text), bg))
	}
	center(s, x, width, y+height-2, "press any key to close", style(theme.Color(theme.Dim), bg))
}

func (a *App) drawPromotion(l layout) {
	s := a.screen
	bg := theme.Color(theme.Panel)
	cell := 3
	if l.large {
		cell = 7
	}
	width := len(promotionOrder)*(cell+2) + 4
	height := 6
	if l.large {
		height = 8
	}
	x := l.boardX + (8*l.cellW-width)/2
	y := l.boardY + (8*l.cellH-height)/2
	titledFrame(s, x, y, width, height, theme.Color(theme.Magenta), bg, "PROMOTE", theme.Color(theme.Magenta))
	color := a.game.Position().Turn
	for i, kind := range promotionOrder {
		px := x + 3 + i*(cell+2)
		square := theme.Color(theme.LightSquare)
		if i == a.promoIndex {
			square = theme.Mix(theme.LightSquare, theme.Cyan, 0.6)
		}
		fill(s, px, y+2, cell, height-5, square)
		drawPiece(s, px, y+2, l.large, chess.Piece{Kind: kind, Color: color}, square)
		key := strings.ToLower(kind.Letter())
		center(s, px, cell, y+height-2, key, style(theme.Color(theme.Lime), bg).Bold(true))
	}
}

func (a *App) drawResult(l layout) {
	s := a.screen
	bg := theme.Color(theme.Panel)
	title, color := "DRAW", theme.Yellow
	if a.game.Status() == chess.Checkmate {
		title, color = "CHECKMATE", theme.Lime
	}
	text, _ := a.statusText()
	width, height := 40, 7
	x := l.boardX + (8*l.cellW-width)/2
	y := l.boardY + (8*l.cellH-height)/2
	titledFrame(s, x, y, width, height, theme.Color(color), bg, title, theme.Color(color))
	center(s, x, width, y+2, text, style(theme.Color(theme.Text), bg).Bold(true))
	hints(s, x, width, y+4, []hint{{"n", "new"}, {"u", "undo"}, {"esc", "view board"}},
		theme.Color(theme.Cyan), theme.Color(theme.Dim), bg)
}
