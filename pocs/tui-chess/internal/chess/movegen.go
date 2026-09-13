package chess

var (
	knightSteps = [][2]int{{1, 2}, {2, 1}, {2, -1}, {1, -2}, {-1, -2}, {-2, -1}, {-2, 1}, {-1, 2}}
	kingSteps   = [][2]int{{1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}, {0, -1}, {1, -1}}
	bishopDirs  = [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	rookDirs    = [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	queenDirs   = append(append([][2]int{}, bishopDirs...), rookDirs...)
	promotions  = []Kind{Queen, Rook, Bishop, Knight}
)

func (p *Position) LegalMoves() []Move {
	pseudo := p.pseudoMoves()
	legal := pseudo[:0]
	for _, m := range pseudo {
		next := p.Play(m)
		if !next.Attacked(next.KingSquare(p.Turn), next.Turn) {
			legal = append(legal, m)
		}
	}
	return legal
}

func (p *Position) Attacked(sq int, by Color) bool {
	if sq == NoSquare {
		return false
	}
	file, rank := FileOf(sq), RankOf(sq)
	pawnRank := rank - forward(by)
	if p.is(file-1, pawnRank, Piece{Pawn, by}) || p.is(file+1, pawnRank, Piece{Pawn, by}) {
		return true
	}
	for _, s := range knightSteps {
		if p.is(file+s[0], rank+s[1], Piece{Knight, by}) {
			return true
		}
	}
	for _, s := range kingSteps {
		if p.is(file+s[0], rank+s[1], Piece{King, by}) {
			return true
		}
	}
	return p.attackedAlong(file, rank, bishopDirs, by, Bishop) || p.attackedAlong(file, rank, rookDirs, by, Rook)
}

func (p *Position) attackedAlong(file, rank int, dirs [][2]int, by Color, kind Kind) bool {
	for _, d := range dirs {
		for x, y := file+d[0], rank+d[1]; onBoard(x, y); x, y = x+d[0], y+d[1] {
			pc := p.Board[Square(x, y)]
			if pc.Empty() {
				continue
			}
			if pc.Color == by && (pc.Kind == kind || pc.Kind == Queen) {
				return true
			}
			break
		}
	}
	return false
}

func (p *Position) pseudoMoves() []Move {
	moves := make([]Move, 0, 64)
	for sq, pc := range p.Board {
		if pc.Empty() || pc.Color != p.Turn {
			continue
		}
		switch pc.Kind {
		case Pawn:
			moves = p.pawnMoves(moves, sq)
		case Knight:
			moves = p.stepMoves(moves, sq, knightSteps)
		case Bishop:
			moves = p.slideMoves(moves, sq, bishopDirs)
		case Rook:
			moves = p.slideMoves(moves, sq, rookDirs)
		case Queen:
			moves = p.slideMoves(moves, sq, queenDirs)
		case King:
			moves = p.stepMoves(moves, sq, kingSteps)
			moves = p.castleMoves(moves, sq)
		}
	}
	return moves
}

func (p *Position) stepMoves(moves []Move, from int, steps [][2]int) []Move {
	file, rank := FileOf(from), RankOf(from)
	for _, s := range steps {
		x, y := file+s[0], rank+s[1]
		if !onBoard(x, y) {
			continue
		}
		to := Square(x, y)
		if target := p.Board[to]; target.Empty() || target.Color != p.Turn {
			moves = append(moves, Move{From: from, To: to})
		}
	}
	return moves
}

func (p *Position) slideMoves(moves []Move, from int, dirs [][2]int) []Move {
	file, rank := FileOf(from), RankOf(from)
	for _, d := range dirs {
		for x, y := file+d[0], rank+d[1]; onBoard(x, y); x, y = x+d[0], y+d[1] {
			to := Square(x, y)
			target := p.Board[to]
			if target.Empty() {
				moves = append(moves, Move{From: from, To: to})
				continue
			}
			if target.Color != p.Turn {
				moves = append(moves, Move{From: from, To: to})
			}
			break
		}
	}
	return moves
}

func (p *Position) pawnMoves(moves []Move, from int) []Move {
	dir := forward(p.Turn)
	file, rank := FileOf(from), RankOf(from)
	startRank, lastRank := 1, 7
	if p.Turn == Black {
		startRank, lastRank = 6, 0
	}
	add := func(to int) {
		if RankOf(to) != lastRank {
			moves = append(moves, Move{From: from, To: to})
			return
		}
		for _, k := range promotions {
			moves = append(moves, Move{From: from, To: to, Promotion: k})
		}
	}
	if onBoard(file, rank+dir) && p.Board[Square(file, rank+dir)].Empty() {
		add(Square(file, rank+dir))
		if rank == startRank && p.Board[Square(file, rank+2*dir)].Empty() {
			add(Square(file, rank+2*dir))
		}
	}
	for _, df := range []int{-1, 1} {
		x, y := file+df, rank+dir
		if !onBoard(x, y) {
			continue
		}
		to := Square(x, y)
		target := p.Board[to]
		if (!target.Empty() && target.Color != p.Turn) || to == p.EnPassant {
			add(to)
		}
	}
	return moves
}

func (p *Position) castleMoves(moves []Move, from int) []Move {
	rank, kingSide, queenSide := 0, WhiteKingSide, WhiteQueenSide
	if p.Turn == Black {
		rank, kingSide, queenSide = 7, BlackKingSide, BlackQueenSide
	}
	enemy := p.Turn.Other()
	if from != Square(4, rank) || p.Castling&(kingSide|queenSide) == 0 || p.Attacked(from, enemy) {
		return moves
	}
	rook := Piece{Rook, p.Turn}
	if p.Castling&kingSide != 0 && p.Board[Square(7, rank)] == rook && p.emptyBetween(rank, 5, 6) &&
		!p.Attacked(Square(5, rank), enemy) && !p.Attacked(Square(6, rank), enemy) {
		moves = append(moves, Move{From: from, To: Square(6, rank)})
	}
	if p.Castling&queenSide != 0 && p.Board[Square(0, rank)] == rook && p.emptyBetween(rank, 1, 3) &&
		!p.Attacked(Square(3, rank), enemy) && !p.Attacked(Square(2, rank), enemy) {
		moves = append(moves, Move{From: from, To: Square(2, rank)})
	}
	return moves
}

func (p *Position) emptyBetween(rank, fromFile, toFile int) bool {
	for f := fromFile; f <= toFile; f++ {
		if !p.Board[Square(f, rank)].Empty() {
			return false
		}
	}
	return true
}
