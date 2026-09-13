package ui

import "github.com/diegopacheco/tui-chess/internal/chess"

const (
	panelWidth     = 36
	panelMinHeight = 20
	columnGap      = 3
)

type layout struct {
	cellW, cellH   int
	boardX, boardY int
	left, top      int
	panelX         int
	height         int
	width          int
	large          bool
	flipped        bool
	fits           bool
}

func newLayout(w, h int, flipped bool) layout {
	var l layout
	for _, size := range [][2]int{{7, 3}, {3, 1}} {
		l = layout{cellW: size[0], cellH: size[1], large: size[0] == 7, flipped: flipped}
		boardW, boardH := 8*l.cellW+4, 8*l.cellH+3
		l.height = max(boardH, panelMinHeight)
		l.width = boardW + columnGap + panelWidth
		l.fits = w >= l.width && h >= l.height+4
		l.left = (w - l.width) / 2
		l.top = (h-l.height-4)/2 + 2
		l.boardX, l.boardY = l.left+3, l.top+1
		l.panelX = l.left + boardW + columnGap
		if l.fits {
			return l
		}
	}
	return l
}

func (l layout) squareOrigin(sq int) (int, int) {
	col, row := chess.FileOf(sq), 7-chess.RankOf(sq)
	if l.flipped {
		col, row = 7-col, 7-row
	}
	return l.boardX + col*l.cellW, l.boardY + row*l.cellH
}

func (l layout) squareAt(x, y int) (int, bool) {
	if x < l.boardX || y < l.boardY {
		return chess.NoSquare, false
	}
	col, row := (x-l.boardX)/l.cellW, (y-l.boardY)/l.cellH
	if col > 7 || row > 7 {
		return chess.NoSquare, false
	}
	if l.flipped {
		col, row = 7-col, 7-row
	}
	return chess.Square(col, 7-row), true
}
