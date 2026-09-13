package chess

import (
	"fmt"
	"strings"
)

type Move struct {
	From      int
	To        int
	Promotion Kind
}

func (m Move) String() string {
	s := SquareName(m.From) + SquareName(m.To)
	if m.Promotion != None {
		s += strings.ToLower(m.Promotion.Letter())
	}
	return s
}

func ParseMove(s string) (Move, error) {
	if len(s) != 4 && len(s) != 5 {
		return Move{}, fmt.Errorf("invalid move %q", s)
	}
	from, err := ParseSquare(s[0:2])
	if err != nil {
		return Move{}, err
	}
	to, err := ParseSquare(s[2:4])
	if err != nil {
		return Move{}, err
	}
	m := Move{From: from, To: to}
	if len(s) == 5 {
		k := strings.IndexByte("nbrq", s[4])
		if k < 0 {
			return Move{}, fmt.Errorf("invalid promotion in %q", s)
		}
		m.Promotion = Knight + Kind(k)
	}
	return m, nil
}
