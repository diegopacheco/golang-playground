package ui

import (
	"strings"
	"time"

	"github.com/diegopacheco/tui-chess/internal/chess"
	"github.com/diegopacheco/tui-chess/internal/engine"
	"github.com/diegopacheco/tui-chess/internal/theme"
	"github.com/gdamore/tcell/v2"
)

type view int

const (
	menuView view = iota
	gameView
)

type opponent int

const (
	computer opponent = iota
	human
)

type computerMove struct {
	generation int
	move       chess.Move
	ok         bool
}

type App struct {
	screen       tcell.Screen
	view         view
	menuIndex    int
	level        int
	opponent     opponent
	player       chess.Color
	game         *chess.Game
	cursor       int
	selected     int
	targets      []chess.Move
	flipped      bool
	help         bool
	promoting    bool
	pending      chess.Move
	promoIndex   int
	resultHidden bool
	thinking     bool
	generation   int
	problem      string
	buttons      tcell.ButtonMask
	quit         bool
}

func New(screen tcell.Screen) *App {
	return &App{screen: screen, level: 1, game: chess.NewGame(), selected: chess.NoSquare}
}

func Run() error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := screen.Init(); err != nil {
		return err
	}
	defer screen.Fini()
	screen.EnableMouse(tcell.MouseButtonEvents)
	screen.HideCursor()
	a := New(screen)
	for !a.quit {
		a.Draw()
		screen.Show()
		a.Handle(screen.PollEvent())
	}
	return nil
}

func (a *App) Draw() {
	s := a.screen
	w, h := s.Size()
	fill(s, 0, 0, w, h, theme.Color(theme.Bg))
	if a.view == menuView {
		a.drawMenu(w, h)
		return
	}
	l := newLayout(w, h, a.flipped)
	if !l.fits {
		drawTooSmall(s, w, h)
		return
	}
	a.drawHeader(l)
	a.drawBoard(l)
	a.drawPanel(l)
	hints(s, 0, w, l.top+l.height+1, []hint{{"←↑↓→", "move"}, {"enter", "select"}, {"u", "undo"}, {"f", "flip"}, {"n", "new"}, {"m", "menu"}, {"?", "help"}, {"q", "quit"}},
		theme.Color(theme.Cyan), theme.Color(theme.Dim), theme.Color(theme.Bg))
	switch {
	case a.help:
		a.drawHelp(w, h)
	case a.promoting:
		a.drawPromotion(l)
	case a.over() && !a.resultHidden:
		a.drawResult(l)
	}
}

func (a *App) drawHeader(l layout) {
	s := a.screen
	bg := theme.Color(theme.Bg)
	y := l.top - 2
	col := l.left + put(s, l.left, y, "♞ ", style(theme.Color(theme.Magenta), bg))
	put(s, col, y, "TUI CHESS", style(theme.Color(theme.Cyan), bg).Bold(true))
	mode := "Two Players"
	if a.opponent == computer {
		mode = "You play " + a.player.String() + " · " + engine.Levels[a.level].Name
	}
	put(s, l.left+l.width-len([]rune(mode)), y, mode, style(theme.Color(theme.Dim), bg))
}

func (a *App) Handle(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventResize:
		a.screen.Sync()
	case *tcell.EventInterrupt:
		if m, ok := ev.Data().(computerMove); ok {
			a.applyComputerMove(m)
		}
	case *tcell.EventKey:
		a.handleKey(ev)
	case *tcell.EventMouse:
		a.handleMouse(ev)
	}
}

func (a *App) handleKey(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyCtrlC {
		a.quit = true
		return
	}
	switch {
	case a.view == menuView:
		a.menuKey(ev)
	case a.help:
		a.help = false
	case a.promoting:
		a.promotionKey(ev)
	default:
		a.gameKey(ev)
	}
}

func isRune(ev *tcell.EventKey, runes string) bool {
	return ev.Key() == tcell.KeyRune && strings.ContainsRune(runes, ev.Rune())
}

func isKey(ev *tcell.EventKey, key tcell.Key, runes string) bool {
	return ev.Key() == key || isRune(ev, runes)
}

func (a *App) menuKey(ev *tcell.EventKey) {
	levels := len(engine.Levels)
	switch {
	case isKey(ev, tcell.KeyUp, "k"):
		a.menuIndex = (a.menuIndex + menuItems - 1) % menuItems
	case isKey(ev, tcell.KeyDown, "j"):
		a.menuIndex = (a.menuIndex + 1) % menuItems
	case isKey(ev, tcell.KeyLeft, "h") && a.menuIndex == itemLevel:
		a.level = (a.level + levels - 1) % levels
	case isKey(ev, tcell.KeyRight, "l") && a.menuIndex == itemLevel:
		a.level = (a.level + 1) % levels
	case isKey(ev, tcell.KeyEnter, " "):
		a.chooseMenuItem()
	case isKey(ev, tcell.KeyEscape, "q"):
		a.quit = true
	}
}

func (a *App) chooseMenuItem() {
	switch a.menuIndex {
	case itemWhite:
		a.newGame(computer, chess.White)
	case itemBlack:
		a.newGame(computer, chess.Black)
	case itemTwoPlayers:
		a.newGame(human, chess.White)
	case itemLevel:
		a.level = (a.level + 1) % len(engine.Levels)
	case itemQuit:
		a.quit = true
	}
}

func (a *App) gameKey(ev *tcell.EventKey) {
	switch {
	case isKey(ev, tcell.KeyUp, "k"):
		a.moveCursor(0, 1)
	case isKey(ev, tcell.KeyDown, "j"):
		a.moveCursor(0, -1)
	case isKey(ev, tcell.KeyLeft, "h"):
		a.moveCursor(-1, 0)
	case isKey(ev, tcell.KeyRight, "l"):
		a.moveCursor(1, 0)
	case isKey(ev, tcell.KeyEnter, " "):
		a.activate(a.cursor)
	case ev.Key() == tcell.KeyEscape && a.over() && !a.resultHidden:
		a.resultHidden = true
	case ev.Key() == tcell.KeyEscape:
		a.clearSelection()
	case isRune(ev, "u"):
		a.undo()
	case isRune(ev, "f"):
		a.flipped = !a.flipped
	case isRune(ev, "n"):
		a.restart()
	case isRune(ev, "m"):
		a.interrupt()
		a.view = menuView
	case isRune(ev, "?"):
		a.help = true
	case isRune(ev, "q"):
		a.quit = true
	}
}

func (a *App) promotionKey(ev *tcell.EventKey) {
	count := len(promotionOrder)
	switch {
	case isKey(ev, tcell.KeyLeft, "h"):
		a.promoIndex = (a.promoIndex + count - 1) % count
	case isKey(ev, tcell.KeyRight, "l"):
		a.promoIndex = (a.promoIndex + 1) % count
	case isKey(ev, tcell.KeyEnter, " "):
		a.promote(a.promoIndex)
	case ev.Key() == tcell.KeyEscape:
		a.promoting = false
	case isRune(ev, "qrbn"):
		a.promote(strings.IndexRune("qrbn", ev.Rune()))
	}
}

func (a *App) handleMouse(ev *tcell.EventMouse) {
	pressed := ev.Buttons()&tcell.ButtonPrimary != 0 && a.buttons&tcell.ButtonPrimary == 0
	a.buttons = ev.Buttons()
	if !pressed || a.view != gameView || a.help || a.promoting {
		return
	}
	w, h := a.screen.Size()
	x, y := ev.Position()
	if sq, ok := newLayout(w, h, a.flipped).squareAt(x, y); ok {
		a.cursor = sq
		a.activate(sq)
	}
}

func (a *App) newGame(opp opponent, player chess.Color) {
	a.view = gameView
	a.opponent, a.player = opp, player
	a.flipped = opp == computer && player == chess.Black
	a.restart()
}

func (a *App) restart() {
	a.interrupt()
	a.game = chess.NewGame()
	a.cursor = chess.Square(4, 1)
	if a.flipped {
		a.cursor = chess.Square(4, 6)
	}
	a.think()
}

func (a *App) interrupt() {
	a.generation++
	a.thinking = false
	a.promoting = false
	a.resultHidden = false
	a.problem = ""
	a.clearSelection()
}

func (a *App) clearSelection() {
	a.selected, a.targets = chess.NoSquare, nil
}

func (a *App) moveCursor(df, dr int) {
	if a.flipped {
		df, dr = -df, -dr
	}
	file := min(max(chess.FileOf(a.cursor)+df, 0), 7)
	rank := min(max(chess.RankOf(a.cursor)+dr, 0), 7)
	a.cursor = chess.Square(file, rank)
}

func (a *App) over() bool {
	return a.game.Status() != chess.Playing
}

func (a *App) humanToMove() bool {
	return !a.over() && !a.thinking && (a.opponent == human || a.game.Position().Turn == a.player)
}

func (a *App) activate(sq int) {
	if !a.humanToMove() {
		return
	}
	for _, m := range a.targets {
		if m.To != sq {
			continue
		}
		if m.Promotion != chess.None {
			a.promoting, a.pending, a.promoIndex = true, m, 0
			return
		}
		a.commit(m)
		return
	}
	pc := a.game.Position().Board[sq]
	if sq == a.selected || pc.Empty() || pc.Color != a.game.Position().Turn {
		a.clearSelection()
		return
	}
	a.selected, a.targets = sq, a.game.MovesFrom(sq)
}

func (a *App) promote(index int) {
	m := a.pending
	m.Promotion = promotionOrder[index]
	a.promoting = false
	a.commit(m)
}

func (a *App) commit(m chess.Move) {
	a.clearSelection()
	if err := a.game.Play(m); err != nil {
		a.problem = err.Error()
		return
	}
	a.problem = ""
	a.think()
}

func (a *App) undo() {
	a.interrupt()
	if a.game.Undo() && a.opponent == computer && a.game.Position().Turn != a.player {
		a.game.Undo()
	}
	a.think()
}

func (a *App) think() {
	if a.view != gameView || a.opponent != computer || a.thinking || a.over() || a.game.Position().Turn == a.player {
		return
	}
	a.thinking = true
	pos, generation, depth := a.game.Position(), a.generation, engine.Levels[a.level].Depth
	go func() {
		m, ok := engine.Best(pos, depth)
		event := tcell.NewEventInterrupt(computerMove{generation: generation, move: m, ok: ok})
		for a.screen.PostEvent(event) != nil {
			time.Sleep(10 * time.Millisecond)
		}
	}()
}

func (a *App) applyComputerMove(m computerMove) {
	if m.generation != a.generation {
		return
	}
	a.thinking = false
	if m.ok {
		a.commit(m.move)
	}
}
