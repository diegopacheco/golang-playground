package chess

import "strings"

type Color uint8

const (
	White Color = iota
	Black
)

func (c Color) Other() Color {
	return c ^ 1
}

func (c Color) String() string {
	if c == White {
		return "White"
	}
	return "Black"
}

type Kind uint8

const (
	None Kind = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)

func (k Kind) Letter() string {
	return string(" PNBRQK"[k])
}

func (k Kind) Points() int {
	return [...]int{0, 1, 3, 3, 5, 9, 0}[k]
}

type Piece struct {
	Kind  Kind
	Color Color
}

func (p Piece) Empty() bool {
	return p.Kind == None
}

func (p Piece) Letter() string {
	if p.Color == Black {
		return strings.ToLower(p.Kind.Letter())
	}
	return p.Kind.Letter()
}
