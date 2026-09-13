package ui

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/diegopacheco/tui-chess/internal/chess"
	"github.com/diegopacheco/tui-chess/internal/theme"
	"github.com/gdamore/tcell/v2"
)

type shot struct {
	name          string
	width, height int
	setup         func(a *App) error
}

var shots = []shot{
	{"01-menu", 110, 30, func(a *App) error {
		return nil
	}},
	{"02-select", 110, 32, func(a *App) error {
		a.newGame(computer, chess.White)
		if err := a.replay("e2e4", "e7e5", "g1f3", "b8c6"); err != nil {
			return err
		}
		a.cursor = chess.Square(5, 0)
		a.activate(a.cursor)
		return nil
	}},
	{"03-check", 110, 32, func(a *App) error {
		a.newGame(computer, chess.White)
		a.level = 2
		if err := a.replay("e2e4", "e7e5", "g1f3", "b8c6", "f1c4", "f8c5", "c2c3", "g8f6", "d2d4", "e5d4", "c3d4", "c5b4"); err != nil {
			return err
		}
		a.cursor = chess.Square(2, 0)
		a.activate(a.cursor)
		return nil
	}},
	{"04-promotion", 110, 32, func(a *App) error {
		a.newGame(human, chess.White)
		if err := a.load("r3k3/1P3ppp/8/8/8/8/5PPP/6K1 w - - 0 30"); err != nil {
			return err
		}
		a.activate(chess.Square(1, 6))
		a.cursor = chess.Square(1, 7)
		a.activate(a.cursor)
		return nil
	}},
	{"05-checkmate", 110, 32, func(a *App) error {
		a.newGame(human, chess.White)
		return a.replay("e2e4", "e7e5", "f1c4", "b8c6", "d1h5", "g8f6", "h5f7")
	}},
	{"06-flipped", 110, 32, func(a *App) error {
		a.newGame(human, chess.White)
		a.opponent, a.player, a.flipped = computer, chess.Black, true
		a.cursor = chess.Square(2, 6)
		return a.replay("d2d4", "g8f6", "c2c4", "e7e6", "b1c3", "f8b4", "e2e3", "e8g8", "f1d3", "d7d5", "g1f3")
	}},
	{"07-help", 110, 32, func(a *App) error {
		a.newGame(human, chess.White)
		a.help = true
		return a.replay("e2e4", "c7c5")
	}},
	{"08-compact", 80, 24, func(a *App) error {
		a.newGame(computer, chess.White)
		if err := a.replay("e2e4", "e7e5", "g1f3", "b8c6", "f1b5", "a7a6", "b5c6", "d7c6", "e1g1"); err != nil {
			return err
		}
		a.thinking = true
		return nil
	}},
}

func (a *App) replay(moves ...string) error {
	for _, s := range moves {
		m, err := chess.ParseMove(s)
		if err != nil {
			return err
		}
		if err := a.game.Play(m); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) load(fen string) error {
	game, err := chess.FromFEN(fen)
	if err != nil {
		return err
	}
	a.game = game
	return nil
}

func Capture(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, sh := range shots {
		page, err := render(sh)
		if err != nil {
			return fmt.Errorf("%s: %w", sh.name, err)
		}
		if err := os.WriteFile(filepath.Join(dir, sh.name+".html"), []byte(page), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func render(sh shot) (string, error) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		return "", err
	}
	defer screen.Fini()
	screen.SetSize(sh.width, sh.height)
	a := New(screen)
	if err := sh.setup(a); err != nil {
		return "", err
	}
	a.Draw()
	screen.Show()
	return screenHTML(screen, sh.name), nil
}

func screenHTML(screen tcell.SimulationScreen, name string) string {
	cells, w, h := screen.GetContents()
	var b strings.Builder
	fmt.Fprintf(&b, `<!doctype html><html><head><meta charset="utf-8"><title>tui-chess %s</title><style>`+
		`body{margin:0;padding:18px;background:%s}`+
		`pre{font:15px/19px 'JetBrains Mono','SF Mono',Menlo,monospace;margin:0;width:max-content;border-radius:10px;overflow:hidden;box-shadow:0 20px 60px #00e5ff22}`+
		`span{white-space:pre;display:inline-block;width:1ch;height:19px;vertical-align:top;overflow:hidden}`+
		`</style></head><body><pre>`, name, theme.Bg)
	for y := range h {
		for x := range w {
			cell := cells[y*w+x]
			fg, bg, attrs := cell.Style.Decompose()
			text := " "
			if len(cell.Runes) > 0 {
				text = string(cell.Runes)
			}
			weight := "400"
			if attrs&tcell.AttrBold != 0 {
				weight = "700"
			}
			fmt.Fprintf(&b, `<span style="color:%s;background:%s;font-weight:%s">%s</span>`,
				cssColor(fg, theme.Text), cssColor(bg, theme.Bg), weight, html.EscapeString(text))
		}
		b.WriteString("\n")
	}
	b.WriteString("</pre></body></html>\n")
	return b.String()
}

func cssColor(c tcell.Color, fallback string) string {
	if hex := c.Hex(); hex >= 0 {
		return fmt.Sprintf("#%06x", hex)
	}
	return fallback
}
