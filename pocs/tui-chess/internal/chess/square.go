package chess

import "fmt"

const NoSquare = -1

func Square(file, rank int) int {
	return rank*8 + file
}

func FileOf(sq int) int {
	return sq % 8
}

func RankOf(sq int) int {
	return sq / 8
}

func SquareName(sq int) string {
	return string([]byte{byte('a' + FileOf(sq)), byte('1' + RankOf(sq))})
}

func ParseSquare(s string) (int, error) {
	if len(s) != 2 || s[0] < 'a' || s[0] > 'h' || s[1] < '1' || s[1] > '8' {
		return NoSquare, fmt.Errorf("invalid square %q", s)
	}
	return Square(int(s[0]-'a'), int(s[1]-'1')), nil
}

func onBoard(file, rank int) bool {
	return file >= 0 && file < 8 && rank >= 0 && rank < 8
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
