package ui

import (
	"fmt"
	"strings"

	"github.com/diegopacheco/tui-chess/internal/chess"
	"github.com/diegopacheco/tui-chess/internal/engine"
	"github.com/diegopacheco/tui-chess/internal/theme"
	"github.com/gdamore/tcell/v2"
)

func (a *App) drawPanel(l layout) {
	s := a.screen
	bg := theme.Color(theme.Panel)
	titledFrame(s, l.panelX, l.top, panelWidth, l.height, theme.Color(theme.Border), bg, "GAME", theme.Color(theme.Cyan))
	x, y := l.panelX+2, l.top+2
	pos := a.game.Position()
	for _, c := range []chess.Color{chess.White, chess.Black} {
		marker := "  "
		if pos.Turn == c && a.game.Status() == chess.Playing {
			marker = "▶ "
		}
		col := x + put(s, x, y, marker, style(theme.Color(theme.Lime), bg))
		s.SetContent(col, y, glyphs[chess.King], nil, style(glyphColor(c), bg))
		col += 2 + put(s, col+2, y, c.String(), style(theme.Color(theme.Text), bg).Bold(true))
		put(s, col+1, y, a.playerName(c), style(theme.Color(theme.Dim), bg))
		y++
	}
	y++
	text, color := a.statusText()
	put(s, x, y, text, style(theme.Color(color), bg).Bold(true))
	y += 2
	y = a.drawCaptured(x, y, bg)
	y++
	put(s, x, y, "MOVES", style(theme.Color(theme.Purple), bg).Bold(true))
	a.drawMoves(x, y+1, l.top+l.height-1, bg)
}

func (a *App) playerName(c chess.Color) string {
	switch {
	case a.opponent == human:
		return fmt.Sprintf("Player %d", c+1)
	case c == a.player:
		return "You"
	}
	return "Computer · " + engine.Levels[a.level].Name
}

func (a *App) statusText() (string, string) {
	pos := a.game.Position()
	switch a.game.Status() {
	case chess.Checkmate:
		return "Checkmate · " + pos.Turn.Other().String() + " wins", theme.Lime
	case chess.Stalemate:
		return "Stalemate · draw", theme.Yellow
	case chess.FiftyMoveRule:
		return "Draw · fifty-move rule", theme.Yellow
	case chess.Repetition:
		return "Draw · threefold repetition", theme.Yellow
	case chess.InsufficientMaterial:
		return "Draw · insufficient material", theme.Yellow
	}
	switch {
	case a.problem != "":
		return a.problem, theme.Red
	case a.thinking:
		return "Computer is thinking…", theme.Purple
	case pos.InCheck():
		return "Check! " + pos.Turn.String() + " to move", theme.Red
	}
	return pos.Turn.String() + " to move", theme.Text
}

func (a *App) drawCaptured(x, y int, bg tcell.Color) int {
	s := a.screen
	put(s, x, y, "CAPTURED", style(theme.Color(theme.Purple), bg).Bold(true))
	scores := [2]int{}
	taken := [2][]chess.Kind{}
	for _, c := range []chess.Color{chess.White, chess.Black} {
		taken[c] = a.game.Captured(c.Other())
		for _, k := range taken[c] {
			scores[c] += k.Points()
		}
	}
	for _, c := range []chess.Color{chess.White, chess.Black} {
		y++
		var b strings.Builder
		for _, k := range taken[c] {
			b.WriteRune(glyphs[k])
		}
		put(s, x, y, c.String(), style(theme.Color(theme.Dim), bg))
		col := x + 7 + put(s, x+7, y, b.String(), style(glyphColor(c.Other()), bg))
		if lead := scores[c] - scores[c.Other()]; lead > 0 {
			put(s, col+1, y, fmt.Sprintf("+%d", lead), style(theme.Color(theme.Lime), bg))
		}
	}
	return y + 1
}

func (a *App) drawMoves(x, y, bottom int, bg tcell.Color) {
	s := a.screen
	sans := a.game.SANs()
	rows := (len(sans) + 1) / 2
	visible := bottom - y
	first := max(0, rows-visible)
	for row := first; row < rows; row++ {
		line := y + row - first
		put(s, x, line, fmt.Sprintf("%3d.", row+1), style(theme.Color(theme.Dim), bg))
		for side := range 2 {
			i := row*2 + side
			if i >= len(sans) {
				break
			}
			color := theme.Text
			if i == len(sans)-1 {
				color = theme.Cyan
			}
			put(s, x+5+side*10, line, sans[i], style(theme.Color(color), bg))
		}
	}
}
