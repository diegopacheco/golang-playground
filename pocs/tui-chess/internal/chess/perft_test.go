package chess

import "testing"

func perft(p *Position, depth int) int {
	if depth == 0 {
		return 1
	}
	moves := p.LegalMoves()
	if depth == 1 {
		return len(moves)
	}
	nodes := 0
	for _, m := range moves {
		next := p.Play(m)
		nodes += perft(&next, depth-1)
	}
	return nodes
}

func TestPerftMatchesKnownNodeCounts(t *testing.T) {
	cases := []struct {
		name  string
		fen   string
		depth int
		nodes int
	}{
		{"start position", StartFEN, 4, 197281},
		{"kiwipete castling pins and en passant", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 3, 97862},
		{"en passant discovered check", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", 5, 674624},
		{"promotions and castling rights loss", "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1", 4, 422333},
		{"talkchess bug catcher", "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", 3, 62379},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, err := ParseFEN(c.fen)
			if err != nil {
				t.Fatal(err)
			}
			if got := perft(&p, c.depth); got != c.nodes {
				t.Fatalf("perft(%d) = %d, want %d: move generation disagrees with the reference engine count", c.depth, got, c.nodes)
			}
		})
	}
}
