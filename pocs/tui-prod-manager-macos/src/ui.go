package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var ink = tcell.GetColor("#d7e0ed")
var muted = tcell.GetColor("#7c8ca6")
var teal = tcell.GetColor("#58d6be")
var paper = tcell.GetColor("#10151f")
var lineColor = tcell.GetColor("#416780")
var tabNames = []string{"Agents", "Processes", "Podman", "Reclaim", "Storage", "AI"}

type UI struct {
	app                              *tview.Application
	pages                            *tview.Pages
	root                             *tview.Flex
	tabs                             *tview.Table
	table                            *tview.Table
	detail, summary, footer          *tview.TextView
	body                             *tview.Flex
	aiForm                           *tview.Form
	modelDrop                        *tview.DropDown
	aiAnswer                         *tview.TextView
	filter                           *tview.InputField
	ctx                              context.Context
	cancel                           context.CancelFunc
	runner                           Runner
	home, scanRoot                   string
	tab                              int
	procs                            []Process
	agents                           []Process
	containers                       []Container
	images                           []Image
	volumes                          []Volume
	podmanMode, procSort             int
	projects                         []Project
	storage                          Storage
	rows                             []any
	busy                             [6]bool
	loaded                           [6]bool
	query                            string
	modal                            bool
	hint, status, statusKind, working string
	statusAt                         time.Time
	announce                         [6]bool
	logPath                          string
	logPID                           int
	logBusy                          bool
	config                           AIConfig
	snapshot                         string
	aiBusy                           bool
	detailVersion                    int
	cancelWork                       [6]context.CancelFunc
	metrics                          Metrics
	cpuHistory                       []float64
	cpuPanel, memoryPanel, diskPanel *tview.TextView
	overview                         *tview.Flex
	header                           *tview.TextView
	previewID                        string
}

func safe(s string) string {
	return tview.Escape(strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || !unicode.IsControl(r) {
			return r
		}
		return -1
	}, s))
}
func textView() *tview.TextView { return tview.NewTextView().SetDynamicColors(true).SetTextColor(ink) }
func panel(p *tview.Box, title string) {
	p.SetBorder(true).SetBorderColor(lineColor).SetTitleColor(teal).SetTitle(" [::b]" + title + "[::-] ").SetTitleAlign(tview.AlignLeft).SetBackgroundColor(paper)
}
func newUI(root string) *UI {
	tview.Borders.TopLeft = '╭'
	tview.Borders.TopRight = '╮'
	tview.Borders.BottomLeft = '╰'
	tview.Borders.BottomRight = '╯'
	tview.Borders.TopLeftFocus = '╭'
	tview.Borders.TopRightFocus = '╮'
	tview.Borders.BottomLeftFocus = '╰'
	tview.Borders.BottomRightFocus = '╯'
	tview.Borders.HorizontalFocus = '─'
	tview.Borders.VerticalFocus = '│'
	tview.Styles.PrimitiveBackgroundColor = paper
	tview.Styles.ContrastBackgroundColor = tcell.GetColor("#172230")
	tview.Styles.MoreContrastBackgroundColor = teal
	tview.Styles.PrimaryTextColor = ink
	tview.Styles.SecondaryTextColor = muted
	tview.Styles.TertiaryTextColor = teal
	tview.Styles.InverseTextColor = tcell.ColorWhite
	tview.Styles.BorderColor = lineColor
	ctx, cancel := context.WithCancel(context.Background())
	home, _ := os.UserHomeDir()
	cfg, configErr := loadConfig()
	u := &UI{app: tview.NewApplication(), ctx: ctx, cancel: cancel, runner: run, home: home, scanRoot: root, config: cfg}
	u.pages = tview.NewPages()
	u.root = tview.NewFlex().SetDirection(tview.FlexRow)
	u.header = textView().SetWrap(false)
	u.cpuPanel = textView().SetWrap(false)
	panel(u.cpuPanel.Box, "cpu")
	u.memoryPanel = textView().SetWrap(false)
	panel(u.memoryPanel.Box, "memory")
	u.memoryPanel.SetBorderColor(tcell.GetColor("#b8a0e4")).SetTitleColor(tcell.GetColor("#b8a0e4"))
	u.diskPanel = textView().SetWrap(false)
	panel(u.diskPanel.Box, "disk")
	u.diskPanel.SetBorderColor(tcell.GetColor("#dcb77a")).SetTitleColor(tcell.GetColor("#dcb77a"))
	u.overview = tview.NewFlex().AddItem(u.cpuPanel, 0, 2, false).AddItem(u.memoryPanel, 0, 1, false).AddItem(u.diskPanel, 0, 1, false)
	u.tabs = tview.NewTable().SetSelectable(false, true).SetSelectedStyle(tcell.StyleDefault.Background(teal).Foreground(paper).Bold(true))
	for i, n := range tabNames {
		u.tabs.SetCell(0, i, tview.NewTableCell(fmt.Sprintf(" %d  %s ", i+1, n)).SetExpansion(1).SetAlign(tview.AlignCenter).SetTextColor(ink))
	}
	u.tabs.SetSelectedFunc(func(row, col int) { u.switchTab(col) })
	u.tabs.SetSelectionChangedFunc(func(row, col int) {
		if u.tab != col {
			u.switchTab(col)
		}
	})
	u.summary = textView()
	u.summary.SetWrap(false).SetTextColor(muted)
	u.summary.SetBorderPadding(0, 0, 1, 1)
	u.filter = tview.NewInputField().SetLabel(" / ").SetFieldBackgroundColor(tcell.GetColor("#172230")).SetFieldTextColor(ink).SetChangedFunc(func(s string) { u.query = s; u.renderTable() })
	u.filter.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			u.filter.SetText("")
		}
		u.app.SetFocus(u.table)
	})
	u.table = tview.NewTable().SetSelectable(true, false).SetFixed(1, 0).SetSelectedStyle(tcell.StyleDefault.Foreground(teal).Background(tcell.GetColor("#223a46")).Bold(true))
	panel(u.table.Box, "processes")
	u.table.SetSelectedFunc(func(row, col int) { u.inspect() })
	u.table.SetSelectionChangedFunc(func(row, col int) { u.preview() })
	u.table.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseLeftClick {
			x, y := event.Position()
			row, col := u.table.CellAt(x, y)
			if row == 0 && u.tab == 1 && col >= 2 && col <= 4 {
				u.procSort = col - 2
				u.renderTable()
				return tview.MouseConsumed, nil
			}
			if row > 0 && row <= len(u.rows) {
				u.table.Select(row, 0)
				u.app.SetFocus(u.table)
				u.inspect()
				return tview.MouseConsumed, nil
			}
		}
		return action, event
	})
	u.detail = textView().SetScrollable(true).SetWrap(true)
	panel(u.detail.Box, "inspector")
	u.detail.SetBorderColor(tcell.GetColor("#58d6be")).SetBorderPadding(1, 1, 1, 1)
	u.detail.SetText("Select a row and press Enter to inspect.\nUse 1–6 or click a tab to navigate.")
	u.layoutBody()
	u.footer = textView()
	u.footer.SetBackgroundColor(tcell.GetColor("#172230"))
	u.root.AddItem(u.header, 1, 0, false).AddItem(u.overview, 6, 0, false).AddItem(u.tabs, 1, 0, false).AddItem(u.summary, 1, 0, false).AddItem(u.body, 0, 1, true).AddItem(u.footer, 2, 0, false)
	u.pages.AddPage("main", u.root, true, true)
	u.buildAI()
	u.app.SetRoot(u.pages, true).EnableMouse(true).SetInputCapture(u.key)
	u.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool { u.updateDashboard(); return false })
	u.renderTable()
	u.updateDashboard()
	if configErr != nil {
		u.detail.SetText("AI configuration: " + safe(configErr.Error()))
	}
	return u
}
func (u *UI) queue(f func()) {
	select {
	case <-u.ctx.Done():
		return
	default:
	}
	u.app.QueueUpdateDraw(func() {
		if u.ctx.Err() == nil {
			f()
		}
	})
}
func (u *UI) switchTab(tab int) {
	if tab < 0 || tab > 5 {
		return
	}
	u.tab = tab
	u.previewID = ""
	u.rows = nil
	u.table.Clear().Select(1, 0).ScrollToBeginning()
	u.logPath = ""
	u.detailVersion++
	u.query = ""
	u.filter.SetText("")
	u.tabs.Select(0, tab)
	u.root.RemoveItem(u.body)
	u.root.RemoveItem(u.footer)
	u.layoutBody()
	u.root.AddItem(u.body, 0, 1, true).AddItem(u.footer, 2, 0, false)
	u.detail.SetText(`[#7c8ca6]Select a row to inspect.

[::b]enter[::-] open details
[::b]a[::-]     explain with AI[-]`)
	u.renderTable()
	if tab == 5 {
		u.app.SetFocus(u.aiForm)
	} else {
		u.app.SetFocus(u.table)
		if !u.loaded[tab] && !u.busy[tab] {
			u.announce[tab] = true
			u.flash("busy", "Loading "+tabNames[tab])
			u.refresh(tab)
		}
	}
}
func (u *UI) key(e *tcell.EventKey) *tcell.EventKey {
	if e.Key() == tcell.KeyCtrlC {
		u.cancel()
		u.app.Stop()
		return nil
	}
	if u.modal {
		return e
	}
	if u.app.GetFocus() == u.filter {
		return e
	}
	if u.tab == 5 && e.Key() == tcell.KeyTab {
		_, button := u.aiForm.GetFocusedItemIndex()
		if u.app.GetFocus() == u.aiAnswer {
			u.app.SetFocus(u.aiForm)
			return nil
		}
		if button == u.aiForm.GetButtonCount()-1 {
			u.app.SetFocus(u.aiAnswer)
			return nil
		}
	}
	if u.tab == 5 && e.Key() != tcell.KeyEscape && e.Key() != tcell.KeyCtrlN && e.Key() != tcell.KeyCtrlP {
		return e
	}
	switch e.Key() {
	case tcell.KeyCtrlN:
		u.switchTab((u.tab + 1) % 6)
		return nil
	case tcell.KeyCtrlP:
		u.switchTab((u.tab + 5) % 6)
		return nil
	case tcell.KeyTab:
		if u.app.GetFocus() == u.detail {
			u.app.SetFocus(u.table)
		} else {
			u.app.SetFocus(u.detail)
		}
		return nil
	case tcell.KeyEscape:
		if u.tab == 5 {
			u.switchTab(0)
		} else {
			u.logPath = ""
			u.detailVersion++
			u.app.SetFocus(u.table)
		}
		return nil
	}
	r := e.Rune()
	if r >= '1' && r <= '6' {
		u.switchTab(int(r - '1'))
		return nil
	}
	switch r {
	case 'q':
		u.cancel()
		u.app.Stop()
	case '/':
		u.app.SetFocus(u.filter)
	case 'r':
		u.manualRefresh()
	case 'c':
		if cancel := u.cancelWork[u.tab]; cancel != nil {
			cancel()
			u.flash("info", "Cancelling "+tabNames[u.tab]+" scan")
		} else {
			u.flash("info", "Nothing is running on "+tabNames[u.tab])
		}
	case 'i':
		u.inspect()
	case 'a':
		u.askSelected()
	case 'm':
		if u.tab == 2 {
			u.podmanMode = (u.podmanMode + 1) % 3
			u.renderTable()
			u.flash("info", "Showing "+[]string{"containers", "images", "volumes"}[u.podmanMode])
		} else {
			u.flash("info", "m switches lists on the Podman tab (3)")
		}
	case 'o':
		if u.tab == 1 {
			u.procSort = (u.procSort + 1) % 3
			u.renderTable()
			u.flash("info", "Sorted by "+sortNames[u.procSort])
		} else {
			u.flash("info", "o sorts the Processes tab (2)")
		}
	case 'x', 'K', 's', 'R', 'e', 'h', 'd', 'S':
		u.action(r)
	case 'l':
		if p, ok := u.selected().(Process); ok && u.tab == 0 {
			u.connect(p, true)
		}
	case '?':
		u.show("Keyboard", "1–6 tabs · Ctrl-N / Ctrl-P next / previous tab\n↑↓ or j/k select · Enter / i inspect · / filter\nTab moves between inventory and inspector · r refresh · c cancel scan\na explain selected item with AI\nProcesses: o or click CPU/MEMORY/AGE sort · filter cpu>5 mem>1GB age>2h · x terminate · K force kill\nPodman: s stop · S start · R restart · K kill · e shell · h SSH · m containers/images/volumes · d force delete image or volume\nReclaim: d reclaim selected project · y confirms any review\nAgents: Enter follow session · l pick any discovered log · wheel or Tab+↑↓ scroll, End resumes live\nAI: Tab cycles fields · Escape returns to Agents\nq / Ctrl-C quit")
	default:
		return e
	}
	return nil
}
func (u *UI) selected() any {
	row, _ := u.table.GetSelection()
	if row < 1 || row > len(u.rows) {
		return nil
	}
	return u.rows[row-1]
}
func rowID(v any) string {
	switch x := v.(type) {
	case Process:
		return fmt.Sprint(x.PID)
	case Container:
		return x.ID
	case Image:
		return x.ID
	case Volume:
		return x.Name
	case Project:
		return x.Path
	case DiskEntry:
		return x.Path
	}
	return ""
}
func (u *UI) renderTable() {
	selected := rowID(u.selected())
	rowOffset, columnOffset := u.table.GetOffset()
	u.table.SetOffset(rowOffset, columnOffset)
	u.rows = nil
	u.table.Clear()
	var headers []string
	var values [][]string
	var objects []any
	summary := ""
	hint := ""
	switch u.tab {
	case 0:
		headers = []string{"AGENT", "PID", "CPU", "RAM"}
		for _, p := range u.agents {
			k := agentKind(p)
			if k == "" {
				k = "Go"
			}
			values = append(values, []string{k, strconv.Itoa(p.PID), p.CPU + "%", bytesLabel(p.Memory)})
			objects = append(objects, p)
		}
		summary = fmt.Sprintf("[::b]Agent observatory[::-]   %d running processes", len(u.agents))
		hint = "Enter follow session   l pick any discovered log   tab / wheel scroll messages   a explain"
	case 1:
		headers = []string{"PID", "USER", "CPU", "MEMORY", "AGE", "COMMAND"}
		headers[2+u.procSort] += " ▼"
		for _, p := range sortProcesses(u.procs, u.procSort) {
			values = append(values, []string{strconv.Itoa(p.PID), p.User, p.CPU + "%", bytesLabel(p.Memory), p.Elapsed, filepath.Base(p.Executable)})
			objects = append(objects, p)
		}
		summary = fmt.Sprintf("[::b]Process inventory[::-]   %d processes · sorted by %s", len(u.procs), sortNames[u.procSort])
		hint = "o / click header sort   / cpu>5 mem>1GB age>2h   i path   x terminate   K kill   a explain"
	case 2:
		switch u.podmanMode {
		case 1:
			headers = []string{"IMAGE ID", "SIZE", "CONTAINERS", "TAGS"}
			for _, im := range u.images {
				tags := strings.Join(im.tags(), ", ")
				if tags == "" {
					tags = "<untagged>"
				}
				values = append(values, []string{shortID(im.ID), bytesLabel(im.Size), strconv.Itoa(im.Containers), tags})
				objects = append(objects, im)
			}
			hint = "m volumes   d force delete image + its containers   i inspect"
		case 2:
			headers = []string{"MOUNTS", "DRIVER", "CREATED", "VOLUME"}
			for _, v := range u.volumes {
				values = append(values, []string{strconv.Itoa(v.MountCount), v.Driver, v.CreatedAt[:min(10, len(v.CreatedAt))], v.Name})
				objects = append(objects, v)
			}
			hint = "m containers   d force delete volume   i inspect"
		default:
			headers = []string{"NAME", "STATE", "STATUS", "IMAGE", "ID"}
			for _, c := range u.containers {
				values = append(values, []string{c.name(), c.State, c.Status, c.Image, shortID(c.ID)})
				objects = append(objects, c)
			}
			hint = "s stop   S start   R restart   K kill   e shell   h SSH   m images / volumes"
		}
		running := 0
		for _, c := range u.containers {
			if c.State == "running" {
				running++
			}
		}
		summary = fmt.Sprintf("[::b]Container harbor[::-]   %d running / %d total · %d images · %d volumes", running, len(u.containers), len(u.images), len(u.volumes))
	case 3:
		headers = []string{"RECLAIMABLE", "ARTIFACTS", "PROJECT"}
		var total int64
		for _, p := range u.projects {
			total += p.Bytes
			size := bytesLabel(p.Bytes)
			if p.Path == u.working {
				size = "deleting…"
			}
			values = append(values, []string{size, strconv.Itoa(len(p.Candidates)), u.shortPath(p.Path)})
			objects = append(objects, p)
		}
		summary = fmt.Sprintf("[::b]Build cleanup[::-]   %s across %d projects", bytesLabel(total), len(u.projects))
		hint = "d review + reclaim project   i inspect paths   r scan again   c cancel"
	case 4:
		headers = []string{"GROUP", "RANK", "SIZE", "SHARE", "PATH"}
		for group, list := range [][]DiskEntry{u.storage.Apps, u.storage.Folders} {
			var total int64
			for _, d := range list {
				total += d.Bytes
			}
			for i, d := range list {
				label := "APP"
				if group == 1 {
					label = "FOLDER"
				}
				pct := 0
				if total > 0 {
					pct = int(d.Bytes * 100 / total)
				}
				values = append(values, []string{label, strconv.Itoa(i + 1), bytesLabel(d.Bytes), strings.Repeat("━", max(1, pct/5)) + fmt.Sprintf(" %d%%", pct), u.shortPath(d.Path)})
				objects = append(objects, d)
			}
		}
		summary = fmt.Sprintf("[::b]Storage atlas[::-]   %s used · %s available · %s capacity", bytesLabel(int64(u.storage.Total-u.storage.Free)), bytesLabel(int64(u.storage.Free)), bytesLabel(int64(u.storage.Total)))
		hint = "Top 10 apps + top 10 home folders   i full path   r rescan   c cancel"
	case 5:
		summary = "[::b]AI workbench[::-]   Your SDK · your CLI authentication"
		hint = "Tab fields   Ctrl-N / Ctrl-P tabs   Esc agents   Ctrl-C quit"
	}
	if u.busy[u.tab] {
		summary += "   [#58d6be]SCANNING…[-]"
	}
	u.summary.SetText(summary)
	u.table.SetTitle(" [::b]" + strings.ToLower(tabNames[u.tab]) + "[::-] ")
	if u.tab == 0 {
		if u.logPath != "" {
			u.detail.SetTitle(" [::b]session messages[::-] · LIVE · wheel / tab+↑↓ scroll · End resumes ")
		} else {
			u.detail.SetTitle(" [::b]session messages[::-] · enter to connect ")
		}
	}
	if u.tab != 0 {
		u.detail.SetTitle(" [::b]inspector[::-] ")
	}
	u.hint = hint
	u.renderFooter()
	for col, h := range headers {
		u.table.SetCell(0, col, tview.NewTableCell(" "+h+" ").SetTextColor(teal).SetBackgroundColor(tcell.GetColor("#172230")).SetAttributes(tcell.AttrBold).SetSelectable(false))
	}
	matchRow := 1
	for i, row := range values {
		search := strings.Join(row, " ") + " " + rowID(objects[i])
		if p, ok := objects[i].(Process); ok {
			search += " " + p.Executable
		}
		if !matchesQuery(u.query, search, objects[i]) {
			continue
		}
		u.rows = append(u.rows, objects[i])
		r := len(u.rows)
		if rowID(objects[i]) == selected {
			matchRow = r
		}
		for col, s := range row {
			cell := tview.NewTableCell(" " + safe(s) + " ").SetTextColor(ink)
			if r%2 == 0 {
				cell.SetBackgroundColor(tcell.GetColor("#121b27"))
			}
			if col == 0 {
				cell.SetTextColor(teal)
			}
			if strings.HasSuffix(s, "%") {
				cell.SetTextColor(tcell.GetColor("#dcb77a"))
			}
			if s == "running" {
				cell.SetTextColor(teal)
			}
			if s == "exited" {
				cell.SetTextColor(muted)
			}
			if col == len(row)-1 {
				cell.SetExpansion(1)
			}
			u.table.SetCell(r, col, cell)
		}
	}
	if len(u.rows) == 0 {
		message := "No matching items"
		if u.busy[u.tab] {
			message = "Collecting local system state…"
		} else if !u.loaded[u.tab] {
			message = "Press r to collect local system state"
		}
		u.table.SetCell(1, 0, tview.NewTableCell(message).SetTextColor(muted).SetSelectable(false))
	} else {
		u.table.Select(matchRow, 0)
	}
}
func (u *UI) refresh(tab int) {
	if tab == 5 || u.busy[tab] {
		return
	}
	u.busy[tab] = true
	u.renderTable()
	duration := 20 * time.Second
	if tab == 3 || tab == 4 {
		duration = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(u.ctx, duration)
	u.cancelWork[tab] = cancel
	go func() {
		defer cancel()
		var err error
		var ps []Process
		var agents []Process
		var cs []Container
		var ims []Image
		var vols []Volume
		var projects []Project
		var warnings []string
		var disk Storage
		last := time.Time{}
		progress := func(path string) {
			if time.Since(last) < 300*time.Millisecond {
				return
			}
			last = time.Now()
			u.queue(func() {
				if u.tab == tab {
					u.summary.SetText("[#58d6be]Scanning…[-] " + safe(path) + "   · c cancel")
				}
			})
		}
		switch tab {
		case 0, 1:
			ps, err = processes(ctx, u.runner)
			if tab == 0 && err == nil {
				for _, p := range ps {
					if agentKind(p) != "" || isGoProcess(p) {
						agents = append(agents, p)
					}
				}
			}
		case 2:
			cs, ims, vols, err = podmanData(ctx, u.runner)
		case 3:
			projects, warnings, err = scanProjects(ctx, u.scanRoot, progress)
		case 4:
			disk = scanStorage(ctx, u.home, progress)
			warnings = disk.Warnings
			err = ctx.Err()
		}
		u.queue(func() {
			u.busy[tab] = false
			u.cancelWork[tab] = nil
			u.loaded[tab] = err == nil
			switch tab {
			case 0:
				u.procs = ps
				u.agents = agents
			case 1:
				u.procs = ps
			case 2:
				u.containers = cs
				u.images = ims
				u.volumes = vols
			case 3:
				u.projects = projects
			case 4:
				u.storage = disk
			}
			if u.announce[tab] {
				u.announce[tab] = false
				if err != nil {
					u.flash("err", tabNames[tab]+" refresh failed: "+firstLine(err.Error()))
				} else {
					u.flash("ok", tabNames[tab]+" refreshed at "+time.Now().Format("15:04:05"))
				}
			}
			if u.tab == tab {
				u.renderTable()
				if err != nil {
					u.detail.SetText("Collection failed or incomplete:\n" + safe(err.Error()))
				} else if len(warnings) > 0 {
					u.detail.SetText("Partial results; inaccessible paths were not fully counted:\n" + safe(strings.Join(warnings, "\n")))
				} else if tab == 3 {
					u.detail.SetText("Root: " + safe(u.scanRoot) + "\nOnly recognized generated folders are counted. Select a project to review all paths.\nAllocated bytes use decimal kB / MB / GB. Actual free space may differ on APFS.")
				} else if tab == 4 {
					u.detail.SetText("Apps: /Applications, /System/Applications, ~/Applications.\nFolders: immediate home directories, sized recursively; hidden folders included.\nShare is relative to the displayed top 10 in each group. APFS clones and hard links can affect totals.")
				}
			}
		})
	}()
}
func (u *UI) manualRefresh() {
	switch {
	case u.tab == 5:
		u.flash("info", "AI has nothing to refresh")
	case u.busy[u.tab]:
		u.flash("info", tabNames[u.tab]+" is already refreshing — c cancels")
	default:
		u.announce[u.tab] = true
		u.flash("busy", "Refreshing "+tabNames[u.tab])
		u.refresh(u.tab)
	}
}
func (u *UI) nope(key rune) {
	u.flash("info", fmt.Sprintf("%c does nothing on this row — press ? for keys", key))
}
func (u *UI) background(title string, refresh bool, task func(context.Context) (string, error)) {
	u.detailVersion++
	version := u.detailVersion
	tab := u.tab
	u.detail.SetText(safe(title) + "…")
	if refresh {
		u.flash("busy", title)
	}
	go func() {
		ctx, cancel := context.WithTimeout(u.ctx, 30*time.Second)
		defer cancel()
		s, err := task(ctx)
		u.queue(func() {
			show := u.tab == tab && u.detailVersion == version
			if refresh {
				if err != nil {
					u.flash("err", title+" failed: "+firstLine(err.Error()))
				} else {
					u.flash("ok", title+" done")
				}
				u.refresh(tab)
			}
			if !show {
				return
			}
			if err != nil {
				s += "\n" + err.Error()
			}
			u.detail.SetText(safe(s)).ScrollToBeginning()
		})
	}()
}
func (u *UI) inspect() {
	item := u.selected()
	if item == nil {
		return
	}
	u.logPath = ""
	switch v := item.(type) {
	case Process:
		if u.tab == 0 {
			u.connect(v, false)
			return
		}
		u.background("Inspecting process", false, func(ctx context.Context) (string, error) { return inspectProcess(ctx, u.runner, v), nil })
	case Container:
		u.background("Inspecting container", false, func(ctx context.Context) (string, error) {
			inspect, err := u.runner(ctx, "podman", "inspect", v.ID)
			if err != nil {
				return inspect, err
			}
			top, e := u.runner(ctx, "podman", "top", v.ID)
			if e != nil {
				top = e.Error()
			}
			return "Processes inside container\n" + top + "\nContainer metadata and network bindings\n" + inspect, nil
		})
	case Image:
		u.background("Inspecting image", false, func(ctx context.Context) (string, error) { return u.runner(ctx, "podman", "image", "inspect", v.ID) })
	case Volume:
		u.background("Inspecting volume", false, func(ctx context.Context) (string, error) {
			return u.runner(ctx, "podman", "volume", "inspect", "--", v.Name)
		})
	case Project:
		var b strings.Builder
		fmt.Fprintf(&b, "%s\n\n", v.Path)
		for _, c := range v.Candidates {
			fmt.Fprintf(&b, "%10s  %s\n", bytesLabel(c.Bytes), c.Path)
		}
		fmt.Fprintf(&b, "\nTotal %s · d reviews permanent deletion", bytesLabel(v.Bytes))
		u.detail.SetText(safe(b.String())).ScrollToBeginning()
	case DiskEntry:
		u.detail.SetText(safe(v.Path + "\n\nAllocated size: " + bytesLabel(v.Bytes) + "\nRead-only storage inventory; use Reclaim for generated artifacts."))
	}
}
func (u *UI) connect(p Process, all bool) {
	u.detailVersion++
	version := u.detailVersion
	u.detail.SetText("Locating session logs…")
	go func() {
		ctx, cancel := context.WithTimeout(u.ctx, 15*time.Second)
		defer cancel()
		sources, err := sessionSources(ctx, u.runner, u.home, p, all)
		u.queue(func() {
			if u.tab != 0 || u.detailVersion != version {
				return
			}
			if err != nil {
				u.detail.SetText(safe(err.Error()))
				return
			}
			if len(sources) == 1 || !all && sources[0].Linked {
				u.follow(sources[0].Path, p.PID)
				return
			}
			u.chooseSource(sources, p.PID)
		})
	}()
}
func (u *UI) chooseSource(sources []LogSource, pid int) {
	list := tview.NewList().ShowSecondaryText(true)
	panel(list.Box, "Choose transcript · Esc cancels")
	for _, s := range sources {
		source := s
		list.AddItem(s.Label+" · "+s.Modified.Format("Jan 02 15:04"), s.Path, 0, func() { u.closeModal(); u.follow(source.Path, pid) })
	}
	list.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyEscape {
			u.closeModal()
			return nil
		}
		return e
	})
	u.openModal(list, 100, 20)
}
func (u *UI) follow(path string, pid int) {
	u.logPath = path
	u.logPID = pid
	u.detailVersion++
	u.detail.SetText("Following " + safe(path) + "…").ScrollToEnd()
	u.pollLog()
}
func (u *UI) pollLog() {
	if u.logPath == "" || u.logBusy || u.tab != 0 {
		return
	}
	u.logBusy = true
	path := u.logPath
	pid := u.logPID
	go func() {
		s, err := tailLog(path)
		u.queue(func() {
			u.logBusy = false
			if u.logPath != path {
				return
			}
			if err != nil {
				s = err.Error()
			}
			u.showLog(pid, path, s)
		})
	}()
}
func (u *UI) showLog(pid int, path, text string) {
	u.detail.SetText(fmt.Sprintf("[#58d6be]LIVE[-] · PID %d · %s\n\n%s", pid, safe(path), safe(text)))
}
func (u *UI) openModal(p tview.Primitive, width, height int) {
	u.modal = true
	center := tview.NewFlex().AddItem(nil, 0, 1, false).AddItem(tview.NewFlex().SetDirection(tview.FlexRow).AddItem(nil, 0, 1, false).AddItem(p, height, 0, true).AddItem(nil, 0, 1, false), width, 0, true).AddItem(nil, 0, 1, false)
	u.pages.AddPage("modal", center, true, true)
	u.app.SetFocus(p)
}
func (u *UI) closeModal() {
	u.pages.RemovePage("modal")
	u.modal = false
	if u.tab == 5 {
		u.app.SetFocus(u.aiForm)
	} else {
		u.app.SetFocus(u.table)
	}
}
func (u *UI) show(title, message string) {
	m := tview.NewModal().SetText(safe(title + "\n\n" + message)).AddButtons([]string{"Close"}).SetDoneFunc(func(int, string) { u.closeModal() })
	u.openModal(m, 90, 20)
}
func (u *UI) confirm(title, message string, action func()) {
	view := textView().SetText(safe(message)).SetScrollable(true)
	view.SetBorderPadding(1, 0, 2, 2)
	cancel := func() {
		u.closeModal()
		u.flash("info", "Cancelled — nothing changed. In the dialog press y to confirm")
	}
	accept := func() { u.closeModal(); action() }
	buttons := tview.NewForm().AddButton("Cancel · Esc", cancel).AddButton("Confirm · y", accept)
	buttons.SetButtonsAlign(tview.AlignCenter)
	setDarkButtons(buttons)
	buttons.SetButtonActivatedStyle(tcell.StyleDefault.Background(tcell.GetColor("#e8878e")).Foreground(paper).Bold(true))
	box := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(view, 0, 1, false).AddItem(buttons, 3, 0, true)
	panel(box.Box, title+" · y confirm · Esc cancel")
	box.SetBorderColor(tcell.GetColor("#e8878e")).SetTitleColor(tcell.GetColor("#e8878e"))
	box.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyEscape || e.Rune() == 'n' {
			cancel()
			return nil
		}
		if e.Rune() == 'y' {
			accept()
			return nil
		}
		if e.Key() == tcell.KeyPgDn {
			view.ScrollToEnd()
			return nil
		}
		if e.Key() == tcell.KeyPgUp {
			view.ScrollToBeginning()
			return nil
		}
		return e
	})
	u.openModal(box, 96, 24)
}
func (u *UI) input(title, label, value string, action func(string)) {
	form := tview.NewForm()
	form.AddInputField(label, value, 60, nil, nil).AddButton("Cancel", func() { u.closeModal() }).AddButton("Apply", func() { s := form.GetFormItem(0).(*tview.InputField).GetText(); u.closeModal(); action(s) })
	setDarkButtons(form)
	panel(form.Box, title)
	form.SetCancelFunc(u.closeModal)
	u.openModal(form, 88, 11)
}
func (u *UI) action(key rune) {
	item := u.selected()
	if item == nil {
		u.flash("info", "Select a row first — this list is empty")
		return
	}
	switch v := item.(type) {
	case Process:
		if key != 'x' && key != 'K' {
			u.nope(key)
			return
		}
		sig := syscall.SIGTERM
		label := "Terminate"
		if key == 'K' {
			sig = syscall.SIGKILL
			label = "Force kill"
		}
		go func() {
			ctx, cancel := context.WithTimeout(u.ctx, 5*time.Second)
			defer cancel()
			identity, err := processIdentity(ctx, u.runner, v.PID)
			u.queue(func() {
				if err != nil {
					u.show("Process unavailable", err.Error())
					return
				}
				u.confirm(label, fmt.Sprintf("%s PID %d?\n%s\n\nUnsaved work may be lost.", label, v.PID, v.Executable), func() {
					u.background(label, true, func(ctx context.Context) (string, error) {
						err := signalProcess(ctx, u.runner, v, identity, sig)
						return label + " requested for PID " + strconv.Itoa(v.PID), err
					})
				})
			})
		}()
	case Container:
		if key == 'e' {
			u.confirm("Container shell", "Open an interactive /bin/sh in "+v.name()+"?", func() { u.external("podman", []string{"exec", "-it", v.ID, "/bin/sh"}) })
			return
		}
		if key == 'h' {
			u.input("SSH to "+v.name(), "SSH user", "root", func(user string) {
				args, err := sshArgs(v, user)
				if err != nil {
					u.show("SSH unavailable", err.Error())
					return
				}
				u.external("ssh", args)
			})
			return
		}
		action := map[rune]string{'s': "stop", 'S': "start", 'R': "restart", 'K': "kill"}[key]
		if action == "" {
			u.nope(key)
			return
		}
		u.confirm("Container "+action, action+" "+v.name()+"?\n"+v.ID, func() {
			u.background(action, true, func(ctx context.Context) (string, error) { return containerAction(ctx, u.runner, v.ID, action) })
		})
	case Image:
		if key != 'd' {
			u.nope(key)
			return
		}
		u.confirm("Force delete image", fmt.Sprintf("Force delete %s?\n%s\n\n%d container(s) using it, running or exited, are removed first.", strings.Join(v.tags(), ", "), v.ID, v.Containers), func() {
			u.background("Deleting image", true, func(ctx context.Context) (string, error) { return containerAction(ctx, u.runner, v.ID, "rmi") })
		})
	case Volume:
		if key != 'd' {
			u.nope(key)
			return
		}
		u.confirm("Force delete volume", fmt.Sprintf("Force delete volume %s?\nDriver %s · %d active mount(s)\n\nIts data is lost and containers using it are removed.", v.Name, v.Driver, v.MountCount), func() {
			u.background("Deleting volume", true, func(ctx context.Context) (string, error) {
				return containerAction(ctx, u.runner, v.Name, "volume-rm")
			})
		})
	case Project:
		if key != 'd' {
			u.nope(key)
			return
		}
		var b strings.Builder
		fmt.Fprintf(&b, "Permanently delete %s of generated files?\n%s\n\n", bytesLabel(v.Bytes), v.Path)
		for _, c := range v.Candidates {
			fmt.Fprintf(&b, "%s  %s\n", bytesLabel(c.Bytes), c.Path)
		}
		b.WriteString("\nGit-tracked or changed folders are skipped. Dependencies must be installed again.\nPgUp/PgDn scroll this review.")
		u.confirm("Reclaim project", b.String(), func() { u.reclaim(v) })
	default:
		u.nope(key)
	}
}
func (u *UI) reclaim(p Project) {
	u.detailVersion++
	u.working = p.Path
	u.renderTable()
	u.detail.SetText("Reclaiming " + safe(p.Path) + "…")
	u.flash("busy", "Reclaiming "+bytesLabel(p.Bytes)+" in "+u.shortPath(p.Path))
	go func() {
		ctx, cancel := context.WithTimeout(u.ctx, 5*time.Minute)
		defer cancel()
		removed, skipped, err := reclaimProject(ctx, u.runner, u.scanRoot, p)
		u.queue(func() {
			var freed int64
			for _, c := range removed {
				freed += c.Bytes
			}
			u.working = ""
			u.forgetCandidates(p.Path, removed)
			switch {
			case err != nil:
				u.flash("err", "Reclaim failed: "+firstLine(err.Error()))
			case len(skipped) > 0:
				u.flash("info", fmt.Sprintf("Freed %s · kept %d tracked folder(s), see inspector", bytesLabel(freed), len(skipped)))
			default:
				u.flash("ok", "Freed "+bytesLabel(freed)+" from "+u.shortPath(p.Path))
			}
			msg := "Removed " + bytesLabel(freed) + " from " + p.Path
			if len(skipped) > 0 {
				msg += "\n\nKept, not safe to delete:\n" + strings.Join(skipped, "\n")
			}
			if err != nil {
				msg += "\n\n" + err.Error()
			}
			if u.tab == 3 {
				u.renderTable()
				u.detail.SetText(safe(msg)).ScrollToBeginning()
			}
		})
	}()
}
func (u *UI) forgetCandidates(path string, removed []Candidate) {
	gone := map[string]bool{}
	for _, c := range removed {
		gone[c.Path] = true
	}
	kept := []Project{}
	for _, p := range u.projects {
		if p.Path == path {
			left := []Candidate{}
			p.Bytes = 0
			for _, c := range p.Candidates {
				if !gone[c.Path] {
					left = append(left, c)
					p.Bytes += c.Bytes
				}
			}
			p.Candidates = left
			if len(left) == 0 {
				continue
			}
		}
		kept = append(kept, p)
	}
	u.projects = kept
}
func (u *UI) external(name string, args []string) {
	u.app.Suspend(func() {
		cmd := exec.CommandContext(u.ctx, name, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			u.detail.SetText(safe(err.Error()))
		}
	})
}
func (u *UI) buildAI() {
	u.aiAnswer = textView().SetScrollable(true)
	panel(u.aiAnswer.Box, "Response")
	u.aiAnswer.SetText("[::b]Understand before you act.[::-]\n\nChoose a provider and select one of its installed models. Existing CLI sign-in is used.\n\nPress a on any inventory row to bring its context here. The review dialog shows exactly what is sent.\n\nYour agents-sdks Go package builds each CLI invocation. Responses explain state; the manager never executes AI-generated commands.")
	u.aiForm = tview.NewForm()
	panel(u.aiForm.Box, "Provider & prompt")
	providers := []string{"claude", "codex", "ollama"}
	current := 0
	for i, p := range providers {
		if p == u.config.Provider {
			current = i
		}
	}
	u.modelDrop = tview.NewDropDown().SetLabel("Model").SetOptions([]string{"loading…"}, nil).SetCurrentOption(0)
	u.aiForm.AddDropDown("Provider", providers, current, func(option string, index int) { u.config.Provider = option; u.loadModels(option) })
	u.aiForm.AddFormItem(u.modelDrop)
	u.aiForm.AddInputField("Question", "Explain this system state", 24, nil, nil)
	u.aiForm.AddButton("Save", func() {
		if err := saveConfig(u.config); err != nil {
			u.show("Configuration", err.Error())
		} else {
			u.show("Configuration", "Saved to "+configPath())
		}
	})
	u.aiForm.AddButton("Review & ask", func() { question := u.aiForm.GetFormItem(2).(*tview.InputField).GetText(); u.askAI(question) })
	u.aiForm.AddButton("Clear", func() { u.snapshot = ""; u.aiAnswer.SetText("Context cleared. Ask a general systems question.") })
	setDarkButtons(u.aiForm)
}
func (u *UI) loadModels(provider string) {
	runner, home := u.runner, u.home
	go func() {
		ctx, cancel := context.WithTimeout(u.ctx, 10*time.Second)
		defer cancel()
		models, err := modelOptions(ctx, runner, home, provider)
		u.queue(func() { u.setModels(provider, models, err) })
	}()
}
func (u *UI) setModels(provider string, models []string, err error) {
	if provider != u.config.Provider {
		return
	}
	if err != nil {
		u.config.Model = ""
		u.modelDrop.SetOptions([]string{"none found"}, nil).SetCurrentOption(0)
		u.aiAnswer.SetText("[#e8878e]Models unavailable for " + safe(provider) + "[-]\n" + safe(err.Error()))
		return
	}
	current := 0
	for i, m := range models {
		if m == u.config.Model {
			current = i
		}
	}
	u.modelDrop.SetOptions(models, func(option string, _ int) { u.config.Model = option }).SetCurrentOption(current)
}
func (u *UI) askSelected() {
	item := u.selected()
	if item == nil {
		return
	}
	var snapshot string
	switch v := item.(type) {
	case Process:
		snapshot = fmt.Sprintf("PID %d\nParent %d\nExecutable %s\nCPU %s%%\nMemory %s\nUser %s", v.PID, v.PPID, v.Executable, v.CPU, bytesLabel(v.Memory), v.User)
	case Container:
		snapshot = fmt.Sprintf("Container %s\nID %s\nImage %s\nState %s\nPorts %+v", v.name(), v.ID, v.Image, v.State, v.Ports)
	case Image:
		snapshot = fmt.Sprintf("Image %s\nTags %v\nSize %s\nContainers %d", v.ID, v.tags(), bytesLabel(v.Size), v.Containers)
	case Volume:
		snapshot = fmt.Sprintf("Volume %s\nDriver %s\nMountpoint %s\nActive mounts %d\nCreated %s", v.Name, v.Driver, v.Mountpoint, v.MountCount, v.CreatedAt)
	case Project:
		snapshot = fmt.Sprintf("Project %s\nReclaimable %s\nFolders %+v", v.Path, bytesLabel(v.Bytes), v.Candidates)
	case DiskEntry:
		snapshot = fmt.Sprintf("Path %s\nAllocated %s", v.Path, bytesLabel(v.Bytes))
	}
	u.snapshot = snapshot
	u.switchTab(5)
	u.aiAnswer.SetText("Collecting selected context…")
	u.aiBusy = true
	go func() {
		ctx, cancel := context.WithTimeout(u.ctx, 10*time.Second)
		defer cancel()
		var extra string
		var err error
		switch v := item.(type) {
		case Process:
			extra, err = u.runner(ctx, "lsof", "-nP", "-a", "-p", strconv.Itoa(v.PID), "-i")
		case Container:
			extra, err = u.runner(ctx, "podman", "inspect", "--format", "{{json .NetworkSettings}}", v.ID)
			top, topErr := u.runner(ctx, "podman", "top", v.ID)
			extra += "\nContainer processes\n" + top
			if topErr != nil {
				extra += "\n" + topErr.Error()
			}
		}
		if err != nil {
			extra += "\nConnection details unavailable: " + err.Error()
		}
		if len(extra) > 24000 {
			extra = extra[:24000] + "\n[truncated]"
		}
		if extra != "" {
			snapshot += "\n\nNetwork / process evidence\n" + extra
		}
		u.queue(func() {
			u.aiBusy = false
			u.snapshot = snapshot
			u.aiAnswer.SetText("Selected context\n\n" + safe(snapshot) + "\n\nReview & ask sends this context with your question.")
		})
	}()

}
func (u *UI) askAI(question string) {
	if u.aiBusy {
		u.show("AI busy", "A request is already running.")
		return
	}
	if err := u.config.validate(); err != nil {
		u.show("AI settings", err.Error())
		return
	}
	config := u.config
	snapshot := u.snapshot
	u.confirm("Send to "+config.Provider, "Model: "+config.Model+"\nQuestion: "+question+"\n\nContext sent to the selected provider:\n"+snapshot+"\n\nPgUp/PgDn scroll. Local paths and process metadata may be included.", func() {
		u.aiBusy = true
		u.aiAnswer.SetText("Waiting for " + safe(config.Provider) + "…  Request limit: 90 seconds.")
		go func() {
			ctx, cancel := context.WithTimeout(u.ctx, 90*time.Second)
			defer cancel()
			answer, err := explain(ctx, u.runner, config, question, snapshot)
			u.queue(func() {
				u.aiBusy = false
				if err != nil {
					answer = err.Error()
				}
				u.aiAnswer.SetText(safe(answer)).ScrollToBeginning()
			})
		}()
	})
}
func (u *UI) start() {
	u.app.SetAfterDrawFunc(func(screen tcell.Screen) { u.app.SetAfterDrawFunc(nil); u.refresh(0); go u.watchMetrics() })
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		count := 0
		for {
			select {
			case <-u.ctx.Done():
				return
			case <-ticker.C:
				u.queue(func() {
					u.pollLog()
					u.renderFooter()
					count++
					if count%3 == 0 && u.tab < 3 && !u.modal {
						u.refresh(u.tab)
					}
				})
			}
		}
	}()
}
