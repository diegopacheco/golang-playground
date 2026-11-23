package card

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/diegopacheco/golang-playground/pocs/lipgloss-huh-fun/internal/config"
)

const width = 60

var (
	accent   = lipgloss.Color("#FF5FD2")
	cyan     = lipgloss.Color("#5FD7FF")
	muted    = lipgloss.Color("#8A8FA8")
	text     = lipgloss.Color("#F2F3F7")
	green    = lipgloss.Color("#5FFFAF")
	red      = lipgloss.Color("#FF6B6B")
	badgeBgs = []string{"#7D56F4", "#2D7FF9", "#00A68C", "#E8871E", "#D6336C"}
)

var (
	frame = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(1, 2).
		Width(width)
	header = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#1A1B26")).
		Background(accent).
		Padding(0, 1)
	title = lipgloss.NewStyle().Bold(true).Foreground(text)
	label = lipgloss.NewStyle().Foreground(muted).Width(10)
	value = lipgloss.NewStyle().Foreground(cyan).Bold(true)
	badge = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1).MarginRight(1)
	ok    = lipgloss.NewStyle().Foreground(green).Bold(true)
	bad   = lipgloss.NewStyle().Foreground(red).Bold(true)
	hint  = lipgloss.NewStyle().Foreground(muted).Italic(true)
)

func Render(c config.Config) string {
	rows := []string{
		lipgloss.JoinHorizontal(lipgloss.Center, header.Render("PROJECT CONFIG"), " ", title.Render(c.Name)),
		"",
		row("name", value.Render(c.Name)),
		row("stack", value.Render(string(c.Stack))),
		row("features", badges(c.Features)),
		"",
		status(c),
	}
	return frame.Render(strings.Join(rows, "\n"))
}

func RenderError(err error) string {
	lines := []string{header.Background(red).Render("INVALID CONFIG"), ""}
	for line := range strings.SplitSeq(err.Error(), "\n") {
		lines = append(lines, bad.Render("✗ ")+lipgloss.NewStyle().Foreground(text).Render(line))
	}
	return frame.BorderForeground(red).Render(strings.Join(lines, "\n"))
}

func row(name, content string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, label.Render(name), content)
}

func badges(features []config.Feature) string {
	if len(features) == 0 {
		return hint.Render("none")
	}
	var parts []string
	for i, f := range features {
		parts = append(parts, badge.Background(lipgloss.Color(badgeBgs[i%len(badgeBgs)])).Render(string(f)))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func status(c config.Config) string {
	if err := c.Validate(); err != nil {
		return bad.Render("✗ invalid")
	}
	return ok.Render("✓ valid") + hint.Render("  ready to scaffold")
}
