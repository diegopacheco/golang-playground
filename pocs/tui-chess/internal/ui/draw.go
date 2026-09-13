package ui

import (
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

func style(fg, bg tcell.Color) tcell.Style {
	return tcell.StyleDefault.Foreground(fg).Background(bg)
}

func put(s tcell.Screen, x, y int, text string, st tcell.Style) int {
	i := 0
	for _, r := range text {
		s.SetContent(x+i, y, r, nil, st)
		i++
	}
	return i
}

func center(s tcell.Screen, x, w, y int, text string, st tcell.Style) {
	put(s, x+(w-utf8.RuneCountInString(text))/2, y, text, st)
}

func fill(s tcell.Screen, x, y, w, h int, bg tcell.Color) {
	st := style(bg, bg)
	for row := y; row < y+h; row++ {
		for col := x; col < x+w; col++ {
			s.SetContent(col, row, ' ', nil, st)
		}
	}
}

func frame(s tcell.Screen, x, y, w, h int, border, bg tcell.Color) {
	fill(s, x, y, w, h, bg)
	st := style(border, bg)
	for col := x + 1; col < x+w-1; col++ {
		s.SetContent(col, y, '─', nil, st)
		s.SetContent(col, y+h-1, '─', nil, st)
	}
	for row := y + 1; row < y+h-1; row++ {
		s.SetContent(x, row, '│', nil, st)
		s.SetContent(x+w-1, row, '│', nil, st)
	}
	s.SetContent(x, y, '╭', nil, st)
	s.SetContent(x+w-1, y, '╮', nil, st)
	s.SetContent(x, y+h-1, '╰', nil, st)
	s.SetContent(x+w-1, y+h-1, '╯', nil, st)
}

func titledFrame(s tcell.Screen, x, y, w, h int, border, bg tcell.Color, title string, titleColor tcell.Color) {
	frame(s, x, y, w, h, border, bg)
	put(s, x+2, y, " "+title+" ", style(titleColor, bg).Bold(true))
}

type hint struct {
	key  string
	text string
}

func hints(s tcell.Screen, x, w, y int, items []hint, keyColor, textColor, bg tcell.Color) {
	total := 0
	for _, h := range items {
		total += utf8.RuneCountInString(h.key) + utf8.RuneCountInString(h.text) + 3
	}
	col := x + (w-total)/2
	for _, h := range items {
		col += put(s, col, y, h.key, style(keyColor, bg).Bold(true))
		col += put(s, col, y, " "+h.text+"  ", style(textColor, bg))
	}
}
