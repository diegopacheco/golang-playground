package chess

type Castling uint8

const (
	WhiteKingSide Castling = 1 << iota
	WhiteQueenSide
	BlackKingSide
	BlackQueenSide
)

type Position struct {
	Board     [64]Piece
	Turn      Color
	Castling  Castling
	EnPassant int
	HalfMove  int
	FullMove  int
}

func forward(c Color) int {
	if c == White {
		return 1
	}
	return -1
}

func rightsLostAt(sq int) Castling {
	switch sq {
	case Square(0, 0):
		return WhiteQueenSide
	case Square(7, 0):
		return WhiteKingSide
	case Square(4, 0):
		return WhiteKingSide | WhiteQueenSide
	case Square(0, 7):
		return BlackQueenSide
	case Square(7, 7):
		return BlackKingSide
	case Square(4, 7):
		return BlackKingSide | BlackQueenSide
	}
	return 0
}

func (p Position) Play(m Move) Position {
	piece := p.Board[m.From]
	moved := piece.Kind
	capture := !p.Board[m.To].Empty()
	p.Board[m.From] = Piece{}
	if moved == Pawn && m.To == p.EnPassant && FileOf(m.From) != FileOf(m.To) {
		p.Board[m.To-8*forward(piece.Color)] = Piece{}
		capture = true
	}
	if moved == King && abs(FileOf(m.To)-FileOf(m.From)) == 2 {
		rank := RankOf(m.From)
		rookFrom, rookTo := Square(7, rank), Square(5, rank)
		if FileOf(m.To) == 2 {
			rookFrom, rookTo = Square(0, rank), Square(3, rank)
		}
		p.Board[rookTo] = p.Board[rookFrom]
		p.Board[rookFrom] = Piece{}
	}
	if m.Promotion != None {
		piece.Kind = m.Promotion
	}
	p.Board[m.To] = piece
	p.Castling &^= rightsLostAt(m.From) | rightsLostAt(m.To)
	p.EnPassant = NoSquare
	if moved == Pawn && abs(m.To-m.From) == 16 && p.enemyPawnBeside(m.To, piece.Color) {
		p.EnPassant = (m.From + m.To) / 2
	}
	if moved == Pawn || capture {
		p.HalfMove = 0
	} else {
		p.HalfMove++
	}
	if p.Turn == Black {
		p.FullMove++
	}
	p.Turn = p.Turn.Other()
	return p
}

func (p *Position) enemyPawnBeside(sq int, c Color) bool {
	enemy := Piece{Kind: Pawn, Color: c.Other()}
	return p.is(FileOf(sq)-1, RankOf(sq), enemy) || p.is(FileOf(sq)+1, RankOf(sq), enemy)
}

func (p *Position) is(file, rank int, want Piece) bool {
	return onBoard(file, rank) && p.Board[Square(file, rank)] == want
}

func (p *Position) KingSquare(c Color) int {
	for sq, pc := range p.Board {
		if pc.Kind == King && pc.Color == c {
			return sq
		}
	}
	return NoSquare
}

func (p *Position) InCheck() bool {
	return p.Attacked(p.KingSquare(p.Turn), p.Turn.Other())
}

func (p *Position) IsCapture(m Move) bool {
	return !p.Board[m.To].Empty() || (p.Board[m.From].Kind == Pawn && m.To == p.EnPassant)
}

func (p *Position) insufficientMaterial() bool {
	minors, knights := 0, 0
	var bishopShades [2]bool
	for sq, pc := range p.Board {
		switch pc.Kind {
		case None, King:
		case Knight:
			minors++
			knights++
		case Bishop:
			minors++
			bishopShades[(FileOf(sq)+RankOf(sq))%2] = true
		default:
			return false
		}
	}
	return minors <= 1 || (knights == 0 && !(bishopShades[0] && bishopShades[1]))
}
