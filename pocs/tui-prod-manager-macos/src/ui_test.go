package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func testUI(t *testing.T) *UI {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	u := newUI(t.TempDir())
	t.Cleanup(u.cancel)
	for i := range u.loaded {
		u.loaded[i] = true
	}
	return u
}
func TestSixTabsRenderAtDifferentSizes(t *testing.T) {
	u := testUI(t)
	u.procs = []Process{{PID: 42, Executable: "/Applications/Long App.app/Contents/MacOS/App", CPU: "3", Memory: 1024}}
	u.agents = u.procs
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()
	for _, size := range [][2]int{{144, 46}, {100, 32}, {80, 24}} {
		s.SetSize(size[0], size[1])
		for tab := range tabNames {
			u.switchTab(tab)
			u.pages.SetRect(0, 0, size[0], size[1])
			u.pages.Draw(s)
			s.Show()
			var all strings.Builder
			for y := 0; y < size[1]; y++ {
				for x := 0; x < size[0]; x++ {
					r, _, _, _ := s.GetContent(x, y)
					all.WriteRune(r)
				}
			}
			if !strings.Contains(all.String(), "H A R B O R") {
				t.Fatalf("header missing on %s at %v", tabNames[tab], size)
			}
		}
	}
}
func TestFilteringAndSelectionAreStable(t *testing.T) {
	u := testUI(t)
	u.procs = []Process{{PID: 11, Executable: "/bin/one"}, {PID: 22, Executable: "/bin/two"}}
	u.switchTab(1)
	u.table.Select(2, 0)
	u.procs = []Process{{PID: 22, Executable: "/bin/two"}, {PID: 11, Executable: "/bin/one"}}
	u.renderTable()
	if u.selected().(Process).PID != 22 {
		t.Fatal("refresh changed selected PID")
	}
	u.filter.SetText("one")
	if len(u.rows) != 1 || u.selected().(Process).PID != 11 {
		t.Fatal(u.rows)
	}
	u.filter.SetText("absent")
	if u.selected() != nil {
		t.Fatal("empty selection points at stale process")
	}
}
func TestConfirmationDefaultsToCancel(t *testing.T) {
	u := testUI(t)
	called := false
	u.confirm("Remove", "Target", func() { called = true })
	if !u.modal {
		t.Fatal("confirmation not shown")
	}
	u.app.GetFocus().InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) { u.app.SetFocus(p) })
	if called || u.modal {
		t.Fatal("default confirmation did not cancel")
	}
}
func TestKeysAndMouseSwitchTabs(t *testing.T) {
	u := testUI(t)
	u.key(tcell.NewEventKey(tcell.KeyRune, '3', tcell.ModNone))
	if u.tab != 2 {
		t.Fatal(u.tab)
	}
	u.tabs.SetRect(0, 0, 120, 2)
	s := tcell.NewSimulationScreen("UTF-8")
	s.Init()
	defer s.Fini()
	s.SetSize(120, 40)
	u.tabs.Draw(s)
	u.tabs.MouseHandler()(tview.MouseLeftClick, tcell.NewEventMouse(25, 0, tcell.Button1, tcell.ModNone), func(p tview.Primitive) { u.app.SetFocus(p) })
	if u.tab != 1 {
		t.Fatalf("mouse selected tab %d", u.tab)
	}
}
func TestTerminalTextCannotInjectStyles(t *testing.T) {
	s := safe("\x1b[31m[red]danger[-]\x00")
	if strings.ContainsRune(s, '\x1b') || strings.ContainsRune(s, '\x00') || s == "[red]danger[-]" {
		t.Fatal(s)
	}
}

func TestSwitchingFromShortInventoryKeepsTopSelectionVisible(t *testing.T) {
	u := testUI(t)
	u.agents = []Process{{PID: 12, Executable: "claude"}}
	for i := 0; i < 200; i++ {
		u.procs = append(u.procs, Process{PID: i + 100, Executable: "/bin/process"})
	}
	screen := tcell.NewSimulationScreen("UTF-8")
	screen.Init()
	defer screen.Fini()
	screen.SetSize(120, 40)
	u.pages.SetRect(0, 0, 120, 40)
	u.switchTab(0)
	u.pages.Draw(screen)
	u.switchTab(1)
	u.pages.Draw(screen)
	offset, _ := u.table.GetOffset()
	row, _ := u.table.GetSelection()
	if offset != 0 || row != 1 {
		t.Fatalf("top process hidden: offset %d selection %d", offset, row)
	}
	u.procs = append(u.procs, Process{PID: 999, Executable: "/bin/new"})
	u.renderTable()
	u.pages.Draw(screen)
	offset, _ = u.table.GetOffset()
	if offset != 0 {
		t.Fatal("refresh followed tail instead of selected process")
	}
}

func TestMouseClickOpensInspector(t *testing.T) {
	u := testUI(t)
	u.projects = []Project{{Path: "/workspace/project", Bytes: 4096}}
	screen := tcell.NewSimulationScreen("UTF-8")
	screen.Init()
	defer screen.Fini()
	screen.SetSize(120, 40)
	u.switchTab(3)
	u.pages.SetRect(0, 0, 120, 40)
	u.pages.Draw(screen)
	x, y, _, _ := u.table.GetInnerRect()
	u.table.MouseHandler()(tview.MouseLeftClick, tcell.NewEventMouse(x+2, y+1, tcell.Button1, tcell.ModNone), func(p tview.Primitive) { u.app.SetFocus(p) })
	if !strings.Contains(u.detail.GetText(true), "/workspace/project") {
		t.Fatal("click did not open selected row")
	}
}

func TestSamplerFailureKeepsLastCPUHistoryVisible(t *testing.T) {
	u := testUI(t)
	u.applyMetrics(Metrics{CPU: 100, Processes: 7})
	u.applyMetrics(Metrics{CPU: 100, Processes: 7})
	u.metrics.Error = "top: signal: killed\nextra"
	u.updateDashboard()
	text := u.cpuPanel.GetText(true)
	if !strings.Contains(text, "100.0%") || !strings.Contains(text, "█") || !strings.Contains(text, "sampler restarting · top: signal: killed") {
		t.Fatal(text)
	}
	u.applyMetrics(Metrics{CPU: 50, Processes: 7})
	if u.metrics.Error != "" || len(u.cpuHistory) != 3 {
		t.Fatal("next live sample did not clear the failure")
	}
}
func TestProcessSortByKeyHeaderClickAndConditions(t *testing.T) {
	u := testUI(t)
	u.procs = []Process{{PID: 1, CPU: "9", Memory: 10, Elapsed: "01:00"}, {PID: 2, CPU: "1", Memory: 900, Elapsed: "3-00:00:00"}, {PID: 3, CPU: "5", Memory: 500, Elapsed: "5-00:00:00"}}
	screen := tcell.NewSimulationScreen("UTF-8")
	screen.Init()
	defer screen.Fini()
	screen.SetSize(140, 40)
	u.pages.SetRect(0, 0, 140, 40)
	u.switchTab(1)
	u.key(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))
	if u.procSort != 1 || u.rows[0].(Process).PID != 2 {
		t.Fatal("o did not sort by memory")
	}
	u.pages.Draw(screen)
	x, y, w, _ := u.table.GetInnerRect()
	for cx := x; cx < x+w; cx++ {
		if row, col := u.table.CellAt(cx, y); row == 0 && col == 4 {
			u.table.MouseHandler()(tview.MouseLeftClick, tcell.NewEventMouse(cx, y, tcell.Button1, tcell.ModNone), func(tview.Primitive) {})
			break
		}
	}
	if u.procSort != 2 || u.rows[0].(Process).PID != 3 {
		t.Fatal("clicking AGE did not sort oldest first")
	}
	if u.procs[0].PID != 1 {
		t.Fatal("sorting the view reordered process data shared with Agents")
	}
	u.filter.SetText("mem>100 age>1d")
	if len(u.rows) != 2 {
		t.Fatal("conditions did not filter", u.rows)
	}
}
func TestModelIsPickedFromDiscoveredList(t *testing.T) {
	u := testUI(t)
	if _, ok := u.aiForm.GetFormItem(1).(*tview.DropDown); !ok {
		t.Fatal("model must be selected, not typed")
	}
	u.config.Provider = "ollama"
	u.config.Model = "llama3.2:latest"
	u.setModels("ollama", []string{"tinyllama:latest", "llama3.2:latest"}, nil)
	if u.config.Model != "llama3.2:latest" {
		t.Fatal("saved model not preselected", u.config.Model)
	}
	u.modelDrop.SetCurrentOption(0)
	if u.config.Model != "tinyllama:latest" {
		t.Fatal("selection not applied")
	}
	u.setModels("codex", []string{"gpt"}, nil)
	if u.config.Model != "tinyllama:latest" {
		t.Fatal("late result for another provider replaced the selection")
	}
	u.setModels("ollama", nil, errors.New("ollama not running"))
	if u.config.validate() == nil {
		t.Fatal("request allowed without an available model")
	}
	dir := t.TempDir()
	t.Setenv("CODEX_HOME", dir)
	put(t, filepath.Join(dir, "models_cache.json"), `{"models":[{"slug":"gpt-6-astra"},{"slug":"gpt-5.5"}]}`)
	codex, err := modelOptions(context.Background(), run, "", "codex")
	if err != nil || !reflect.DeepEqual(codex, []string{"gpt-6-astra", "gpt-5.5"}) {
		t.Fatal(codex, err)
	}
	ollama, err := modelOptions(context.Background(), func(context.Context, string, ...string) (string, error) {
		return "NAME ID SIZE MODIFIED\nllama3.2:latest a80c 2.0 GB 7 weeks ago\n", nil
	}, "", "ollama")
	if err != nil || !reflect.DeepEqual(ollama, []string{"llama3.2:latest"}) {
		t.Fatal(ollama, err)
	}
}
func TestLiveMessagesKeepReaderScrollPosition(t *testing.T) {
	u := testUI(t)
	screen := tcell.NewSimulationScreen("UTF-8")
	screen.Init()
	defer screen.Fini()
	screen.SetSize(120, 40)
	u.pages.SetRect(0, 0, 120, 40)
	u.switchTab(0)
	lines := func(n int) string {
		var b strings.Builder
		for i := 0; i < n; i++ {
			fmt.Fprintf(&b, "message %d\n", i)
		}
		return b.String()
	}
	path := filepath.Join(t.TempDir(), "session.log")
	u.follow(path, 7)
	u.showLog(7, path, lines(200))
	u.pages.Draw(screen)
	end, _ := u.detail.GetScrollOffset()
	if end == 0 {
		t.Fatal("new session did not open at the latest message")
	}
	x, y, _, _ := u.detail.GetInnerRect()
	for i := 0; i < 5; i++ {
		u.detail.MouseHandler()(tview.MouseScrollUp, tcell.NewEventMouse(x+1, y+1, tcell.WheelUp, tcell.ModNone), func(tview.Primitive) {})
	}
	u.pages.Draw(screen)
	reading, _ := u.detail.GetScrollOffset()
	u.showLog(7, path, lines(260))
	u.pages.Draw(screen)
	if now, _ := u.detail.GetScrollOffset(); now != reading || reading != end-5 {
		t.Fatalf("new messages yanked the reader from line %d to %d", reading, now)
	}
	u.detail.InputHandler()(tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone), func(tview.Primitive) {})
	u.showLog(7, path, lines(300))
	u.pages.Draw(screen)
	if now, _ := u.detail.GetScrollOffset(); now <= reading+60 {
		t.Fatal("End did not resume following new messages")
	}
}
func TestReclaimUpdatesListAndYConfirms(t *testing.T) {
	u := testUI(t)
	u.projects = []Project{{Path: "/p/a", Bytes: 30, Candidates: []Candidate{{"/p/a/node_modules", 10}, {"/p/a/dist", 20}}}, {Path: "/p/b", Bytes: 5, Candidates: []Candidate{{"/p/b/target", 5}}}}
	u.forgetCandidates("/p/a", []Candidate{{"/p/a/node_modules", 10}})
	if len(u.projects) != 2 || u.projects[0].Bytes != 20 || len(u.projects[0].Candidates) != 1 {
		t.Fatal("reclaimed folder still listed as reclaimable", u.projects)
	}
	u.forgetCandidates("/p/b", []Candidate{{"/p/b/target", 0}})
	if len(u.projects) != 1 {
		t.Fatal("fully reclaimed project still listed")
	}
	called := false
	u.confirm("Reclaim", "Target", func() { called = true })
	u.pages.InputHandler()(tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone), func(p tview.Primitive) { u.app.SetFocus(p) })
	if !called || u.modal {
		t.Fatal("y did not confirm the reviewed action")
	}
}
func TestEveryActionGivesVisibleFeedback(t *testing.T) {
	u := testUI(t)
	u.runner = func(context.Context, string, ...string) (string, error) { return "", nil }
	footer := func() string { return u.footer.GetText(true) }
	press := func(r rune) { u.key(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)) }
	u.switchTab(1)
	press('d')
	if !strings.Contains(footer(), "Select a row first") {
		t.Fatal("d on an empty list was silent:", footer())
	}
	u.procs = []Process{{PID: 5, Executable: "/bin/x"}}
	u.renderTable()
	press('d')
	if !strings.Contains(footer(), "d does nothing on this row") {
		t.Fatal("inapplicable key was silent:", footer())
	}
	press('r')
	if !strings.Contains(footer(), "Refreshing Processes") {
		t.Fatal("r gave no signal:", footer())
	}
	press('r')
	if !strings.Contains(footer(), "already refreshing") {
		t.Fatal("second r was silently ignored:", footer())
	}
	u.confirm("Reclaim", "Target", func() {})
	u.app.GetFocus().InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) { u.app.SetFocus(p) })
	if !strings.Contains(footer(), "Cancelled — nothing changed") {
		t.Fatal("cancelling a review was silent:", footer())
	}
	u.flash("busy", "Scanning Storage")
	u.statusAt = time.Now().Add(-12 * time.Second)
	u.renderFooter()
	if !strings.Contains(footer(), "Scanning Storage · 12s") {
		t.Fatal("long scan shows no progress:", footer())
	}
	u.flash("ok", "old news")
	u.statusAt = time.Now().Add(-9 * time.Second)
	u.renderFooter()
	if strings.Contains(footer(), "old news") {
		t.Fatal("finished message never gives the hint line back")
	}
}
func TestReclaimShowsRowBeingDeleted(t *testing.T) {
	u := testUI(t)
	u.runner = func(context.Context, string, ...string) (string, error) { return "", nil }
	p := Project{Path: "/nowhere/proj", Bytes: 4096, Candidates: []Candidate{{"/nowhere/proj/node_modules", 4096}}}
	u.projects = []Project{p}
	u.switchTab(3)
	u.reclaim(p)
	if u.table.GetCell(1, 0).Text != " deleting… " || !strings.Contains(u.footer.GetText(true), "Reclaiming 4.1 kB") {
		t.Fatal("reclaim started without a visible signal:", u.table.GetCell(1, 0).Text, u.footer.GetText(true))
	}
}
func TestCompactCommandRowsStillFilterFullPaths(t *testing.T) {
	u := testUI(t)
	u.procs = []Process{{PID: 42, Executable: "/Applications/Harbor Tool.app/Contents/MacOS/tool"}}
	u.switchTab(1)
	u.filter.SetText("Harbor Tool.app")
	if len(u.rows) != 1 {
		t.Fatal("full executable path no longer searchable")
	}
}
