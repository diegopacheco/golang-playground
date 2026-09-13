package chess

import (
	"errors"
	"fmt"
	"slices"
)

type Status int

const (
	Playing Status = iota
	Checkmate
	Stalemate
	FiftyMoveRule
	Repetition
	InsufficientMaterial
)

type Game struct {
	positions []Position
	moves     []Move
	sans      []string
}

func NewGame() *Game {
	g, err := FromFEN(StartFEN)
	if err != nil {
		panic(err)
	}
	return g
}

func FromFEN(fen string) (*Game, error) {
	p, err := ParseFEN(fen)
	if err != nil {
		return nil, err
	}
	return &Game{positions: []Position{p}}, nil
}

func (g *Game) Position() Position {
	return g.positions[len(g.positions)-1]
}

func (g *Game) SANs() []string {
	return g.sans
}

func (g *Game) LastMove() (Move, bool) {
	if len(g.moves) == 0 {
		return Move{}, false
	}
	return g.moves[len(g.moves)-1], true
}

func (g *Game) MovesFrom(sq int) []Move {
	pos := g.Position()
	var moves []Move
	for _, m := range pos.LegalMoves() {
		if m.From == sq {
			moves = append(moves, m)
		}
	}
	return moves
}

func (g *Game) Play(m Move) error {
	if g.Status() != Playing {
		return errors.New("game is over")
	}
	pos := g.Position()
	if !slices.Contains(pos.LegalMoves(), m) {
		return fmt.Errorf("illegal move %s", m)
	}
	g.sans = append(g.sans, pos.SAN(m))
	g.moves = append(g.moves, m)
	g.positions = append(g.positions, pos.Play(m))
	return nil
}

func (g *Game) Undo() bool {
	n := len(g.moves)
	if n == 0 {
		return false
	}
	g.moves = g.moves[:n-1]
	g.sans = g.sans[:n-1]
	g.positions = g.positions[:n]
	return true
}

func (g *Game) Status() Status {
	pos := g.Position()
	if len(pos.LegalMoves()) == 0 {
		if pos.InCheck() {
			return Checkmate
		}
		return Stalemate
	}
	switch {
	case pos.HalfMove >= 100:
		return FiftyMoveRule
	case g.repetitions() >= 3:
		return Repetition
	case pos.insufficientMaterial():
		return InsufficientMaterial
	}
	return Playing
}

func (g *Game) repetitions() int {
	key := g.Position().key()
	count := 0
	for _, p := range g.positions {
		if p.key() == key {
			count++
		}
	}
	return count
}

func (g *Game) Captured(c Color) []Kind {
	var kinds []Kind
	for i, m := range g.moves {
		pos := g.positions[i]
		if !pos.IsCapture(m) || pos.Turn == c {
			continue
		}
		kind := pos.Board[m.To].Kind
		if kind == None {
			kind = Pawn
		}
		kinds = append(kinds, kind)
	}
	slices.Sort(kinds)
	return kinds
}
