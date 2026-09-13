package ui

import (
	"fmt"

	"github.com/diegopacheco/tui-chess/internal/chess"
	"github.com/diegopacheco/tui-chess/internal/engine"
	"github.com/diegopacheco/tui-chess/internal/theme"
	"github.com/gdamore/tcell/v2"
)

const (
	itemWhite = iota
	itemBlack
	itemTwoPlayers
	itemLevel
	itemQuit
	menuItems
)

const menuWidth = 34

var banner = []string{
	" █████  ██  ██  ██████   █████   █████ ",
	"██      ██  ██  ██      ██      ██     ",
	"██      ██████  █████    ████    ████  ",
	"██      ██  ██  ██          ██      ██ ",
	" █████  ██  ██  ██████  █████   █████  ",
}

var parade = []chess.Piece{
	{Kind: chess.Rook, Color: chess.White},
	{Kind: chess.Knight, Color: chess.Black},
	{Kind: chess.Bishop, Color: chess.White},
	{Kind: chess.Queen, Color: chess.Black},
	{Kind: chess.King, Color: chess.White},
	{Kind: chess.Pawn, Color: chess.Black},
}

func (a *App) menuLabel(item int) string {
	switch item {
	case itemWhite:
		return "Play White vs Computer"
	case itemBlack:
		return "Play Black vs Computer"
	case itemTwoPlayers:
		return "Two Players"
	case itemLevel:
		return "Difficulty  ◂ " + engine.Levels[a.level].Name + " ▸"
	}
	return "Quit"
}

func (a *App) drawMenu(w, h int) {
	s := a.screen
	bg := theme.Color(theme.Bg)
	showParade := h >= 26
	total := len(banner) + 2 + menuItems + 4
	if showParade {
		total += 4
	}
	y := max(0, (h-total)/2)
	bannerW := len([]rune(banner[0]))
	x := (w - bannerW) / 2
	for row, line := range banner {
		for col, r := range []rune(line) {
			tint := theme.Mix(theme.Cyan, theme.Magenta, float64(col)/float64(bannerW))
			s.SetContent(x+col, y+row, r, nil, style(tint, bg))
		}
	}
	y += len(banner) + 1
	center(s, 0, w, y, "terminal chess · written in go", style(theme.Color(theme.Dim), bg))
	y += 2
	if showParade {
		a.drawParade(w, y)
		y += 4
	}
	a.drawMenuItems(w, y)
	y += menuItems + 2
	hints(s, 0, w, y, []hint{{"↑↓", "choose"}, {"←→", "difficulty"}, {"enter", "start"}, {"q", "quit"}},
		theme.Color(theme.Cyan), theme.Color(theme.Dim), bg)
}

func (a *App) drawParade(w, y int) {
	x := (w - len(parade)*7) / 2
	for i, pc := range parade {
		square := theme.Color(theme.DarkSquare)
		if i%2 == 0 {
			square = theme.Color(theme.LightSquare)
		}
		fill(a.screen, x+i*7, y, 7, 3, square)
		drawPiece(a.screen, x+i*7, y, true, pc, square)
	}
}

func (a *App) drawMenuItems(w, y int) {
	s := a.screen
	x := (w - menuWidth) / 2
	bg := theme.Color(theme.Bg)
	for item := range menuItems {
		rowBg, fg, marker := bg, theme.Color(theme.Text), "  "
		if item == a.menuIndex {
			rowBg, fg, marker = theme.Color(theme.Border), theme.Color(theme.Cyan), "▸ "
		}
		fill(s, x, y+item, menuWidth, 1, rowBg)
		st := style(fg, rowBg).Bold(item == a.menuIndex)
		put(s, x+1, y+item, marker+a.menuLabel(item), st)
	}
}

func drawTooSmall(s tcell.Screen, w, h int) {
	bg := theme.Color(theme.Bg)
	center(s, 0, w, h/2-1, "terminal too small", style(theme.Color(theme.Red), bg).Bold(true))
	l := newLayout(0, 0, false)
	msg := fmt.Sprintf("resize to at least %dx%d", l.width, l.height+4)
	center(s, 0, w, h/2+1, msg, style(theme.Color(theme.Dim), bg))
}
