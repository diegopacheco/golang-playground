package chess

import "strings"

func (p *Position) SAN(m Move) string {
	pc := p.Board[m.From]
	var b strings.Builder
	if pc.Kind == King && abs(FileOf(m.To)-FileOf(m.From)) == 2 {
		if FileOf(m.To) == 6 {
			b.WriteString("O-O")
		} else {
			b.WriteString("O-O-O")
		}
	} else {
		capture := p.IsCapture(m)
		if pc.Kind == Pawn {
			if capture {
				b.WriteByte(SquareName(m.From)[0])
			}
		} else {
			b.WriteString(pc.Kind.Letter())
			b.WriteString(p.disambiguation(m))
		}
		if capture {
			b.WriteByte('x')
		}
		b.WriteString(SquareName(m.To))
		if m.Promotion != None {
			b.WriteString("=" + m.Promotion.Letter())
		}
	}
	next := p.Play(m)
	if next.InCheck() {
		if len(next.LegalMoves()) == 0 {
			b.WriteByte('#')
		} else {
			b.WriteByte('+')
		}
	}
	return b.String()
}

func (p *Position) disambiguation(m Move) string {
	kind := p.Board[m.From].Kind
	rival, sameFile, sameRank := false, false, false
	for _, o := range p.LegalMoves() {
		if o.To != m.To || o.From == m.From || p.Board[o.From].Kind != kind {
			continue
		}
		rival = true
		sameFile = sameFile || FileOf(o.From) == FileOf(m.From)
		sameRank = sameRank || RankOf(o.From) == RankOf(m.From)
	}
	name := SquareName(m.From)
	switch {
	case !rival:
		return ""
	case !sameFile:
		return name[:1]
	case !sameRank:
		return name[1:]
	default:
		return name
	}
}
