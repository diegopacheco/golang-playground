package engine

import (
	"slices"

	"github.com/diegopacheco/tui-chess/internal/chess"
)

const (
	infinity  = 1_000_000
	mateScore = 100_000
)

type Level struct {
	Name  string
	Depth int
}

var Levels = []Level{
	{Name: "Easy", Depth: 1},
	{Name: "Medium", Depth: 3},
	{Name: "Hard", Depth: 4},
}

func Best(p chess.Position, depth int) (chess.Move, bool) {
	moves := p.LegalMoves()
	if len(moves) == 0 {
		return chess.Move{}, false
	}
	order(&p, moves)
	best, alpha := moves[0], -infinity
	for _, m := range moves {
		next := p.Play(m)
		if score := -search(&next, depth-1, -infinity, -alpha, 1); score > alpha {
			best, alpha = m, score
		}
	}
	return best, true
}

func search(p *chess.Position, depth, alpha, beta, ply int) int {
	moves := p.LegalMoves()
	if len(moves) == 0 {
		if p.InCheck() {
			return -mateScore + ply
		}
		return 0
	}
	if p.HalfMove >= 100 {
		return 0
	}
	if depth <= 0 {
		return quiesce(p, moves, alpha, beta)
	}
	order(p, moves)
	for _, m := range moves {
		next := p.Play(m)
		score := -search(&next, depth-1, -beta, -alpha, ply+1)
		if score >= beta {
			return beta
		}
		alpha = max(alpha, score)
	}
	return alpha
}

func quiesce(p *chess.Position, moves []chess.Move, alpha, beta int) int {
	stand := Evaluate(p)
	if stand >= beta {
		return beta
	}
	alpha = max(alpha, stand)
	tactical := slices.DeleteFunc(moves, func(m chess.Move) bool {
		return !p.IsCapture(m) && m.Promotion == chess.None
	})
	order(p, tactical)
	for _, m := range tactical {
		next := p.Play(m)
		score := -quiesce(&next, next.LegalMoves(), -beta, -alpha)
		if score >= beta {
			return beta
		}
		alpha = max(alpha, score)
	}
	return alpha
}

func order(p *chess.Position, moves []chess.Move) {
	slices.SortStableFunc(moves, func(a, b chess.Move) int {
		return priority(p, b) - priority(p, a)
	})
}

func priority(p *chess.Position, m chess.Move) int {
	score := pieceValue[m.Promotion]
	if p.IsCapture(m) {
		victim := max(pieceValue[p.Board[m.To].Kind], pieceValue[chess.Pawn])
		score += 10*victim - pieceValue[p.Board[m.From].Kind]
	}
	return score
}
