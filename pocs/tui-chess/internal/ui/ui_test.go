package ui

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/diegopacheco/tui-chess/internal/chess"
	"github.com/gdamore/tcell/v2"
)

func newTestApp(t *testing.T, w, h int) (*App, tcell.SimulationScreen) {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(screen.Fini)
	screen.SetSize(w, h)
	a := New(screen)
	a.level = 0
	return a, screen
}

func key(r rune) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
}

func click(a *App, sq int) {
	w, h := a.screen.Size()
	x, y := newLayout(w, h, a.flipped).squareOrigin(sq)
	a.Handle(tcell.NewEventMouse(x+1, y, tcell.ButtonPrimary, tcell.ModNone))
	a.Handle(tcell.NewEventMouse(x+1, y, tcell.ButtonNone, tcell.ModNone))
}

func square(t *testing.T, name string) int {
	t.Helper()
	sq, err := chess.ParseSquare(name)
	if err != nil {
		t.Fatal(err)
	}
	return sq
}

func waitForComputer(t *testing.T, a *App, screen tcell.SimulationScreen) {
	t.Helper()
	for a.thinking {
		a.Handle(screen.PollEvent())
	}
}

func TestClicksLandOnTheSquareThatIsDrawn(t *testing.T) {
	for _, size := range [][2]int{{120, 40}, {80, 24}} {
		for _, flipped := range []bool{false, true} {
			l := newLayout(size[0], size[1], flipped)
			if !l.fits {
				t.Fatalf("%dx%d should fit a board", size[0], size[1])
			}
			for sq := range 64 {
				x, y := l.squareOrigin(sq)
				corner := [2]int{x + l.cellW - 1, y + l.cellH - 1}
				for _, p := range [][2]int{{x, y}, corner} {
					if got, ok := l.squareAt(p[0], p[1]); !ok || got != sq {
						t.Fatalf("%dx%d flipped=%v: click at %v hit %d, want %s", size[0], size[1], flipped, p, got, chess.SquareName(sq))
					}
				}
			}
		}
	}
}

func TestFlippedBoardPutsBlackAtTheBottom(t *testing.T) {
	l := newLayout(120, 40, true)
	_, a1y := l.squareOrigin(chess.Square(0, 0))
	_, a8y := l.squareOrigin(chess.Square(0, 7))
	if a8y <= a1y {
		t.Fatal("flipped board must draw rank 8 below rank 1 so black plays from the bottom")
	}
}

func TestTwoPlayersMoveWithTheMouse(t *testing.T) {
	a, _ := newTestApp(t, 120, 40)
	a.newGame(human, chess.White)
	click(a, square(t, "e2"))
	if len(a.targets) != 2 {
		t.Fatalf("e2 pawn shows %d targets, want 2", len(a.targets))
	}
	click(a, square(t, "e4"))
	click(a, square(t, "e7"))
	click(a, square(t, "e5"))
	if got := a.game.SANs(); !slices.Equal(got, []string{"e4", "e5"}) {
		t.Fatalf("moves = %v, want both players to have moved", got)
	}
}

func TestSelectingAnOpponentPieceDoesNothing(t *testing.T) {
	a, _ := newTestApp(t, 120, 40)
	a.newGame(human, chess.White)
	click(a, square(t, "e7"))
	if a.selected != chess.NoSquare || len(a.targets) != 0 {
		t.Fatal("white was allowed to select a black pawn")
	}
}

func TestKeyboardPromotionChoosesThePiece(t *testing.T) {
	a, _ := newTestApp(t, 120, 40)
	a.newGame(human, chess.White)
	if err := a.load("8/1P2k3/8/8/8/8/8/4K3 w - - 0 1"); err != nil {
		t.Fatal(err)
	}
	a.cursor = square(t, "b7")
	a.Handle(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	a.Handle(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone))
	a.Handle(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	if !a.promoting {
		t.Fatal("moving a pawn to the last rank must ask which piece to promote to")
	}
	a.Handle(key('n'))
	if got := a.game.Position().Board[square(t, "b8")]; got != (chess.Piece{Kind: chess.Knight, Color: chess.White}) {
		t.Fatalf("b8 holds %+v, want the knight the player chose", got)
	}
}

func TestHumanCannotMoveDuringComputerTurn(t *testing.T) {
	a, screen := newTestApp(t, 120, 40)
	a.newGame(computer, chess.Black)
	if !a.thinking {
		t.Fatal("computer plays white and must start thinking right away")
	}
	click(a, square(t, "e7"))
	if a.selected != chess.NoSquare {
		t.Fatal("black selected a piece while the computer was on move")
	}
	waitForComputer(t, a, screen)
	if len(a.game.SANs()) != 1 || a.game.Position().Turn != chess.Black {
		t.Fatalf("after the computer moved: moves=%v turn=%v", a.game.SANs(), a.game.Position().Turn)
	}
}

func TestUndoAgainstComputerGivesTheTurnBack(t *testing.T) {
	a, screen := newTestApp(t, 120, 40)
	a.newGame(computer, chess.White)
	click(a, square(t, "e2"))
	click(a, square(t, "e4"))
	waitForComputer(t, a, screen)
	if len(a.game.SANs()) != 2 {
		t.Fatalf("moves = %v, want the player move and the computer reply", a.game.SANs())
	}
	a.Handle(key('u'))
	if len(a.game.SANs()) != 0 || a.game.Position().Turn != chess.White {
		t.Fatal("undo must remove the computer reply and the player move so the player moves again")
	}
}

func TestStaleComputerMoveIsIgnoredAfterUndo(t *testing.T) {
	a, _ := newTestApp(t, 120, 40)
	a.newGame(computer, chess.White)
	click(a, square(t, "e2"))
	click(a, square(t, "e4"))
	stale := a.generation
	a.Handle(key('u'))
	m, err := chess.ParseMove("e7e5")
	if err != nil {
		t.Fatal(err)
	}
	a.Handle(tcell.NewEventInterrupt(computerMove{generation: stale, move: m, ok: true}))
	if len(a.game.SANs()) != 0 {
		t.Fatalf("a reply computed for an undone position was played: %v", a.game.SANs())
	}
}

func TestCaptureWritesEveryShot(t *testing.T) {
	dir := t.TempDir()
	if err := Capture(dir); err != nil {
		t.Fatal(err)
	}
	for _, sh := range shots {
		info, err := os.Stat(filepath.Join(dir, sh.name+".html"))
		if err != nil || info.Size() == 0 {
			t.Fatalf("%s was not written, the README printscreens depend on it: %v", sh.name, err)
		}
	}
}
