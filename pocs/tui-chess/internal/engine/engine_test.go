package engine

import (
	"slices"
	"testing"
	"time"

	"github.com/diegopacheco/tui-chess/internal/chess"
)

func position(t *testing.T, fen string) chess.Position {
	t.Helper()
	p, err := chess.ParseFEN(fen)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func bestAt(t *testing.T, fen string, level Level) string {
	t.Helper()
	m, ok := Best(position(t, fen), level.Depth)
	if !ok {
		t.Fatalf("%s: no move found", level.Name)
	}
	return m.String()
}

func TestEveryLevelDeliversMateInOne(t *testing.T) {
	for _, level := range Levels {
		if got := bestAt(t, "6k1/5ppp/8/8/8/8/5PPP/R5K1 w - - 0 1", level); got != "a1a8" {
			t.Errorf("%s played %s, missing the back-rank mate Ra8#", level.Name, got)
		}
	}
}

func TestEveryLevelTakesAHangingQueen(t *testing.T) {
	for _, level := range Levels {
		if got := bestAt(t, "4k3/8/8/3q4/8/8/8/3QK3 w - - 0 1", level); got != "d1d5" {
			t.Errorf("%s played %s instead of winning the free queen", level.Name, got)
		}
	}
}

func TestDeeperLevelsAvoidPoisonedCaptures(t *testing.T) {
	fen := "3rk3/8/8/3p4/8/8/8/3QK3 w - - 0 1"
	for _, level := range Levels[1:] {
		if got := bestAt(t, fen, level); got == "d1d5" {
			t.Errorf("%s gave the queen for a pawn defended by the d8 rook", level.Name)
		}
	}
}

func TestBlackEngineDefendsAgainstMate(t *testing.T) {
	fen := "r1bqkbnr/pppp1ppp/2n5/4p2Q/2B1P3/8/PPPP1PPP/RNB1K1NR b KQkq - 3 3"
	p := position(t, fen)
	for _, level := range Levels[1:] {
		m, _ := Best(p, level.Depth)
		next := p.Play(m)
		for _, reply := range next.LegalMoves() {
			after := next.Play(reply)
			if after.InCheck() && len(after.LegalMoves()) == 0 {
				t.Fatalf("%s played %s and allowed scholar's mate with %s", level.Name, m, reply)
			}
		}
	}
}

func TestBestReportsNoMoveWhenGameIsOver(t *testing.T) {
	if _, ok := Best(position(t, "7k/5Q2/6K1/8/8/8/8/8 b - - 0 1"), 3); ok {
		t.Fatal("engine returned a move in a stalemate")
	}
}

func TestEvaluationIsSymmetric(t *testing.T) {
	start := position(t, chess.StartFEN)
	if got := Evaluate(&start); got != 0 {
		t.Fatalf("start position scored %d, a balanced position must score 0", got)
	}
	up := position(t, "4k3/8/8/8/8/8/8/3QK3 w - - 0 1")
	down := position(t, "4k3/8/8/8/8/8/8/3QK3 b - - 0 1")
	if Evaluate(&up) <= 0 || Evaluate(&down) != -Evaluate(&up) {
		t.Fatalf("queen-up side must score positive and the other side the exact negative: %d %d", Evaluate(&up), Evaluate(&down))
	}
}

func TestHardLevelAnswersQuicklyInAnOpenMiddlegame(t *testing.T) {
	p := position(t, "r1bq1rk1/pp2bppp/2n1pn2/3p4/2PP4/2N1PN2/PP3PPP/R2QKB1R w KQ - 0 9")
	hard := Levels[len(Levels)-1]
	start := time.Now()
	m, ok := Best(p, hard.Depth)
	if !ok || !slices.Contains(p.LegalMoves(), m) {
		t.Fatalf("hard level returned an illegal move %s", m)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("hard level took %s, a player would think the game froze", elapsed)
	}
}
