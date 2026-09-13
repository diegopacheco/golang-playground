package ui

import (
	"github.com/diegopacheco/tui-chess/internal/chess"
	"github.com/diegopacheco/tui-chess/internal/theme"
	"github.com/gdamore/tcell/v2"
)

var sprites = [7][3]string{
	chess.Pawn:   {"  ▄  ", " ▐█▌ ", " ▟█▙ "},
	chess.Knight: {" ▄▟█▖", "▝▘▐█▌", " ▟██▙"},
	chess.Bishop: {" ▗▆▖ ", "  █  ", " ▟█▙ "},
	chess.Rook:   {"█▄█▄█", " ▐█▌ ", "▟███▙"},
	chess.Queen:  {"▙▗█▖▟", " ▜█▛ ", "▟███▙"},
	chess.King:   {" ▄█▄ ", "  █  ", "▟███▙"},
}

var glyphs = [7]rune{
	chess.Pawn:   '♟',
	chess.Knight: '♞',
	chess.Bishop: '♝',
	chess.Rook:   '♜',
	chess.Queen:  '♛',
	chess.King:   '♚',
}

func pieceColor(c chess.Color) tcell.Color {
	if c == chess.White {
		return theme.Color(theme.WhitePiece)
	}
	return theme.Color(theme.BlackPiece)
}

func glyphColor(c chess.Color) tcell.Color {
	if c == chess.White {
		return theme.Color(theme.WhitePiece)
	}
	return theme.Color(theme.BlackGlyph)
}

func drawPiece(s tcell.Screen, x, y int, large bool, pc chess.Piece, bg tcell.Color) {
	st := style(pieceColor(pc.Color), bg)
	if !large {
		s.SetContent(x+1, y, glyphs[pc.Kind], nil, st.Bold(true))
		return
	}
	for row, line := range sprites[pc.Kind] {
		put(s, x+1, y+row, line, st)
	}
}
