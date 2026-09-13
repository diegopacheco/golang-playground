package chess

import (
	"slices"
	"testing"
)

func play(t *testing.T, g *Game, moves ...string) {
	t.Helper()
	for _, s := range moves {
		m, err := ParseMove(s)
		if err != nil {
			t.Fatal(err)
		}
		if err := g.Play(m); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
	}
}

func TestFENRoundTripKeepsEveryField(t *testing.T) {
	fen := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R b Kq e3 3 17"
	p, err := ParseFEN(fen)
	if err != nil {
		t.Fatal(err)
	}
	if got := p.FEN(); got != fen {
		t.Fatalf("FEN() = %q, want %q", got, fen)
	}
}

func TestParseFENRejectsPositionsTheEngineCannotPlay(t *testing.T) {
	for _, fen := range []string{
		"8/8/8/8/8/8/8/8 w - - 0 1",
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP w KQkq - 0 1",
		"rnbqkbnr/pppppppp/9/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR x KQkq - 0 1",
	} {
		if _, err := ParseFEN(fen); err == nil {
			t.Errorf("ParseFEN(%q) accepted an invalid position", fen)
		}
	}
}

func TestFoolsMateEndsTheGameForBlack(t *testing.T) {
	g := NewGame()
	play(t, g, "f2f3", "e7e5", "g2g4", "d8h4")
	if g.Status() != Checkmate {
		t.Fatalf("status = %v, want checkmate", g.Status())
	}
	if last := g.SANs()[3]; last != "Qh4#" {
		t.Fatalf("mating move SAN = %q, want Qh4#", last)
	}
	if err := g.Play(Move{From: Square(4, 1), To: Square(4, 3)}); err == nil {
		t.Fatal("a finished game accepted another move")
	}
}

func TestStalemateIsADrawNotAWin(t *testing.T) {
	g, err := FromFEN("7k/5Q2/6K1/8/8/8/8/8 b - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	if g.Status() != Stalemate {
		t.Fatalf("status = %v, want stalemate", g.Status())
	}
}

func TestIllegalMovesAreRejected(t *testing.T) {
	g, err := FromFEN("4k3/8/8/8/8/8/4r3/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	capture := Move{From: Square(4, 0), To: Square(4, 1)}
	if err := g.Play(capture); err != nil {
		t.Fatalf("capturing the checking rook should be legal: %v", err)
	}
	g, _ = FromFEN("4k3/4r3/8/8/8/8/4B3/4K3 w - - 0 1")
	if err := g.Play(Move{From: Square(4, 1), To: Square(3, 2)}); err == nil {
		t.Fatal("moving a pinned bishop exposed the king and was accepted")
	}
}

func TestCastlingMovesTheRookAndDropsRights(t *testing.T) {
	g, err := FromFEN("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	play(t, g, "e1g1", "e8c8")
	p := g.Position()
	if p.Board[Square(5, 0)] != (Piece{Rook, White}) || !p.Board[Square(7, 0)].Empty() {
		t.Fatal("white king-side castle did not move the h1 rook to f1")
	}
	if p.Board[Square(3, 7)] != (Piece{Rook, Black}) || !p.Board[Square(0, 7)].Empty() {
		t.Fatal("black queen-side castle did not move the a8 rook to d8")
	}
	if p.Castling != 0 {
		t.Fatalf("castling rights = %b after both kings castled, want none", p.Castling)
	}
	if got := g.SANs(); !slices.Equal(got, []string{"O-O", "O-O-O"}) {
		t.Fatalf("SAN = %v", got)
	}
}

func TestCastlingThroughAttackedSquareIsIllegal(t *testing.T) {
	g, err := FromFEN("4kr2/8/8/8/8/8/8/4K2R w K - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Play(Move{From: Square(4, 0), To: Square(6, 0)}); err == nil {
		t.Fatal("king castled across f1 which the f8 rook attacks")
	}
}

func TestEnPassantRemovesThePassedPawn(t *testing.T) {
	g := NewGame()
	play(t, g, "e2e4", "a7a6", "e4e5", "d7d5", "e5d6")
	p := g.Position()
	if !p.Board[Square(3, 4)].Empty() {
		t.Fatal("en passant left the captured pawn on d5")
	}
	if got := g.SANs()[4]; got != "exd6" {
		t.Fatalf("en passant SAN = %q, want exd6", got)
	}
	if got := g.Captured(Black); !slices.Equal(got, []Kind{Pawn}) {
		t.Fatalf("captured black pieces = %v, want one pawn", got)
	}
}

func TestPromotionReplacesThePawn(t *testing.T) {
	g, err := FromFEN("8/1P2k3/8/8/8/8/8/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	if n := len(g.MovesFrom(Square(1, 6))); n != 4 {
		t.Fatalf("pawn on b7 has %d moves, want 4 promotion choices", n)
	}
	play(t, g, "b7b8n")
	if g.Position().Board[Square(1, 7)] != (Piece{Knight, White}) {
		t.Fatal("under-promotion did not produce a knight")
	}
	if got := g.SANs()[0]; got != "b8=N" {
		t.Fatalf("SAN = %q, want b8=N", got)
	}
}

func TestSANDisambiguatesTwinPieces(t *testing.T) {
	g, err := FromFEN("4k3/8/8/8/8/8/8/R4RK1 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	play(t, g, "a1d1")
	if got := g.SANs()[0]; got != "Rad1" {
		t.Fatalf("SAN = %q, want Rad1 because both rooks reach d1", got)
	}
}

func TestThreefoldRepetitionIsADraw(t *testing.T) {
	g := NewGame()
	play(t, g, "g1f3", "g8f6", "f3g1", "f6g8", "g1f3", "g8f6", "f3g1")
	if g.Status() != Playing {
		t.Fatal("draw declared after only two repetitions")
	}
	play(t, g, "f6g8")
	if g.Status() != Repetition {
		t.Fatalf("status = %v, want repetition", g.Status())
	}
}

func TestInsufficientMaterial(t *testing.T) {
	cases := map[string]Status{
		"4k3/8/8/8/8/8/8/4K3 w - - 0 1":    InsufficientMaterial,
		"4k3/8/8/8/8/8/8/2B1K3 w - - 0 1":  InsufficientMaterial,
		"4k3/8/8/8/8/8/8/1NN1K3 w - - 0 1": Playing,
		"4k3/8/8/8/8/8/4P3/4K3 w - - 0 1":  Playing,
	}
	for fen, want := range cases {
		g, err := FromFEN(fen)
		if err != nil {
			t.Fatal(err)
		}
		if got := g.Status(); got != want {
			t.Errorf("%s: status = %v, want %v", fen, got, want)
		}
	}
}

func TestFiftyMoveRule(t *testing.T) {
	g, err := FromFEN("4k3/8/8/8/8/8/8/R3K3 w - - 99 80")
	if err != nil {
		t.Fatal(err)
	}
	play(t, g, "a1a2")
	if g.Status() != FiftyMoveRule {
		t.Fatalf("status = %v, want fifty-move draw", g.Status())
	}
}

func TestUndoRestoresThePreviousPosition(t *testing.T) {
	g := NewGame()
	before := g.Position().FEN()
	play(t, g, "e2e4")
	if !g.Undo() {
		t.Fatal("undo failed")
	}
	if got := g.Position().FEN(); got != before {
		t.Fatalf("after undo FEN = %q, want %q", got, before)
	}
	if g.Undo() {
		t.Fatal("undo succeeded with no moves played")
	}
}
