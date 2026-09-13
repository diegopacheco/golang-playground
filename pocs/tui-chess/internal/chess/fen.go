package chess

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const StartFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

const castlingLetters = "KQkq"

func ParseFEN(fen string) (Position, error) {
	fields := strings.Fields(fen)
	if len(fields) != 4 && len(fields) != 6 {
		return Position{}, fmt.Errorf("fen needs 4 or 6 fields: %q", fen)
	}
	p := Position{EnPassant: NoSquare, FullMove: 1}
	if err := p.parsePlacement(fields[0]); err != nil {
		return Position{}, err
	}
	switch fields[1] {
	case "w":
		p.Turn = White
	case "b":
		p.Turn = Black
	default:
		return Position{}, fmt.Errorf("invalid side to move %q", fields[1])
	}
	if fields[2] != "-" {
		for _, ch := range fields[2] {
			i := strings.IndexRune(castlingLetters, ch)
			if i < 0 {
				return Position{}, fmt.Errorf("invalid castling rights %q", fields[2])
			}
			p.Castling |= Castling(1 << i)
		}
	}
	if fields[3] != "-" {
		sq, err := ParseSquare(fields[3])
		if err != nil {
			return Position{}, err
		}
		p.EnPassant = sq
	}
	if len(fields) == 6 {
		var err error
		if p.HalfMove, err = strconv.Atoi(fields[4]); err != nil {
			return Position{}, fmt.Errorf("invalid halfmove clock %q", fields[4])
		}
		if p.FullMove, err = strconv.Atoi(fields[5]); err != nil {
			return Position{}, fmt.Errorf("invalid fullmove number %q", fields[5])
		}
	}
	return p, nil
}

func (p *Position) parsePlacement(placement string) error {
	rows := strings.Split(placement, "/")
	if len(rows) != 8 {
		return fmt.Errorf("fen placement needs 8 ranks: %q", placement)
	}
	kings := [2]int{}
	for i, row := range rows {
		rank, file := 7-i, 0
		for _, ch := range row {
			if ch >= '1' && ch <= '8' {
				file += int(ch - '0')
				continue
			}
			kind := strings.IndexRune("pnbrqk", unicode.ToLower(ch))
			if kind < 0 || file > 7 {
				return fmt.Errorf("invalid fen rank %q", row)
			}
			color := White
			if unicode.IsLower(ch) {
				color = Black
			}
			pc := Piece{Kind: Pawn + Kind(kind), Color: color}
			if pc.Kind == King {
				kings[color]++
			}
			p.Board[Square(file, rank)] = pc
			file++
		}
		if file != 8 {
			return fmt.Errorf("fen rank %q does not have 8 files", row)
		}
	}
	if kings != [2]int{1, 1} {
		return fmt.Errorf("fen needs exactly one king per side: %q", placement)
	}
	return nil
}

func (p Position) FEN() string {
	return fmt.Sprintf("%s %d %d", p.key(), p.HalfMove, p.FullMove)
}

func (p Position) key() string {
	var b strings.Builder
	for rank := 7; rank >= 0; rank-- {
		empty := 0
		for file := 0; file < 8; file++ {
			pc := p.Board[Square(file, rank)]
			if pc.Empty() {
				empty++
				continue
			}
			if empty > 0 {
				b.WriteString(strconv.Itoa(empty))
				empty = 0
			}
			b.WriteString(pc.Letter())
		}
		if empty > 0 {
			b.WriteString(strconv.Itoa(empty))
		}
		if rank > 0 {
			b.WriteByte('/')
		}
	}
	b.WriteString(" " + strings.ToLower(p.Turn.String()[:1]) + " ")
	rights := ""
	for i := range castlingLetters {
		if p.Castling&Castling(1<<i) != 0 {
			rights += castlingLetters[i : i+1]
		}
	}
	if rights == "" {
		rights = "-"
	}
	b.WriteString(rights + " ")
	if p.EnPassant == NoSquare {
		b.WriteString("-")
	} else {
		b.WriteString(SquareName(p.EnPassant))
	}
	return b.String()
}
