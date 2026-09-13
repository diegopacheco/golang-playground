package ui

import (
	"slices"

	"github.com/diegopacheco/tui-chess/internal/chess"
	"github.com/diegopacheco/tui-chess/internal/theme"
	"github.com/gdamore/tcell/v2"
)

func (a *App) drawBoard(l layout) {
	s := a.screen
	bg := theme.Color(theme.Bg)
	frame(s, l.boardX-1, l.boardY-1, 8*l.cellW+2, 8*l.cellH+2, theme.Color(theme.Border), bg)
	pos := a.game.Position()
	for sq := range 64 {
		x, y := l.squareOrigin(sq)
		square := a.squareColor(&pos, sq)
		fill(s, x, y, l.cellW, l.cellH, square)
		pc := pos.Board[sq]
		switch {
		case !pc.Empty():
			drawPiece(s, x, y, l.large, pc, square)
		case a.isTarget(sq):
			s.SetContent(x+l.cellW/2, y+l.cellH/2, '●', nil, style(theme.Color(theme.Cyan), square))
		}
	}
	labels := style(theme.Color(theme.Dim), bg)
	for i := range 8 {
		file, rank := i, 7-i
		if l.flipped {
			file, rank = 7-i, i
		}
		s.SetContent(l.boardX+i*l.cellW+l.cellW/2, l.boardY+8*l.cellH+1, rune('a'+file), nil, labels)
		s.SetContent(l.boardX-3, l.boardY+i*l.cellH+l.cellH/2, rune('1'+rank), nil, labels)
	}
}

func (a *App) isTarget(sq int) bool {
	return slices.ContainsFunc(a.targets, func(m chess.Move) bool { return m.To == sq })
}

func (a *App) squareColor(pos *chess.Position, sq int) tcell.Color {
	base := theme.DarkSquare
	if (chess.FileOf(sq)+chess.RankOf(sq))%2 == 1 {
		base = theme.LightSquare
	}
	last, played := a.game.LastMove()
	switch {
	case sq == a.selected:
		return theme.Mix(base, theme.Cyan, 0.6)
	case pos.InCheck() && sq == pos.KingSquare(pos.Turn):
		return theme.Mix(base, theme.Red, 0.75)
	case sq == a.cursor:
		return theme.Mix(base, theme.Lime, 0.5)
	case a.isTarget(sq) && !pos.Board[sq].Empty():
		return theme.Mix(base, theme.Magenta, 0.55)
	case played && (sq == last.From || sq == last.To):
		return theme.Mix(base, theme.Yellow, 0.4)
	}
	return theme.Color(base)
}
