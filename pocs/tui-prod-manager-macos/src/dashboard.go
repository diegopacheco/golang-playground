package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (u *UI) layoutBody() {
	if u.tab == 5 {
		u.body = tview.NewFlex().AddItem(u.aiForm, 42, 0, true).AddItem(u.aiAnswer, 0, 1, false)
		return
	}
	inventory := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(u.filter, 1, 0, false).AddItem(u.table, 0, 1, true)
	left, right := 5, 3
	if u.tab == 0 {
		left, right = 3, 5
	}
	u.body = tview.NewFlex().AddItem(inventory, 0, left, true).AddItem(u.detail, 0, right, false)
}
func (u *UI) shortPath(path string) string {
	if rel, err := filepath.Rel(u.scanRoot, path); err == nil && rel != ".." && !strings.HasPrefix(rel, "../") {
		return rel
	}
	if strings.HasPrefix(path, u.home+"/") {
		return "~/" + strings.TrimPrefix(path, u.home+"/")
	}
	return path
}
func (u *UI) preview() {
	item := u.selected()
	if item == nil || u.detail == nil {
		return
	}
	id := rowID(item)
	if id == u.previewID {
		return
	}
	u.previewID = id
	u.logPath = ""
	u.detailVersion++
	var b strings.Builder
	heading := func(label, value string) { fmt.Fprintf(&b, "[#7c8ca6]%s[-]\n[::b]%s[::-]\n\n", label, safe(value)) }
	switch v := item.(type) {
	case Process:
		heading("PROCESS", filepath.Base(v.Executable))
		fmt.Fprintf(&b, "[#58d6be]pid[-] %-10d [#58d6be]parent[-] %d\n[#58d6be]cpu[-] %-10s [#58d6be]memory[-] %s\n[#58d6be]user[-] %s\n[#58d6be]age[-] %s\n\n", v.PID, v.PPID, v.CPU+"%", bytesLabel(v.Memory), safe(v.User), safe(v.Elapsed))
		heading("EXECUTABLE", v.Executable)
		if u.tab == 0 {
			b.WriteString("[#58d6be]enter[-]  Connect to session messages\n[#58d6be]l[-]      Pick any discovered log\n[#58d6be]a[-]      Ask AI about this process\n\n[#7c8ca6]Claude / Codex transcripts and redirected\nGo output update here every two seconds.[-]")
		} else {
			b.WriteString("[#58d6be]i[-]  Full path, command & connections\n[#dcb77a]x[-]  Terminate    [#e8878e]K[-] Force kill\n[#58d6be]a[-]  Explain with AI")
		}
	case Container:
		heading("CONTAINER", v.name())
		fmt.Fprintf(&b, "[#58d6be]state[-]  %s\n[#58d6be]status[-] %s\n\n", safe(v.State), safe(v.Status))
		heading("IMAGE", v.Image)
		heading("ID", v.ID)
		b.WriteString("[#7c8ca6]PUBLISHED PORTS[-]\n")
		if len(v.Ports) == 0 {
			b.WriteString("None\n")
		}
		for _, p := range v.Ports {
			fmt.Fprintf(&b, "[#58d6be]%d[-] → %d/%s\n", p.HostPort, p.ContainerPort, safe(p.Protocol))
		}
		b.WriteString("\n[#58d6be]i[-] inspect   [#58d6be]e[-] shell   [#58d6be]h[-] SSH\n[#58d6be]m[-] images and volumes")
	case Image:
		heading("IMAGE", strings.Join(v.tags(), "\n"))
		heading("SIZE", bytesLabel(v.Size))
		heading("ID", v.ID)
		heading("CONTAINERS USING IT", fmt.Sprint(v.Containers))
		b.WriteString("[#58d6be]i[-] inspect   [#e8878e]d[-] force delete image + its containers")
	case Volume:
		heading("VOLUME", v.Name)
		heading("DRIVER", v.Driver)
		heading("MOUNTPOINT", v.Mountpoint)
		heading("CREATED", v.CreatedAt)
		heading("ACTIVE MOUNTS", fmt.Sprint(v.MountCount))
		b.WriteString("[#58d6be]i[-] inspect   [#e8878e]d[-] force delete volume")
	case Project:
		heading("PROJECT", u.shortPath(v.Path))
		heading("RECLAIMABLE", bytesLabel(v.Bytes))
		b.WriteString("[#7c8ca6]GENERATED ARTIFACTS[-]\n\n")
		for _, c := range v.Candidates {
			fmt.Fprintf(&b, "[#58d6be]%s[-]  %s\n", bytesLabel(c.Bytes), safe(strings.TrimPrefix(c.Path, v.Path+"/")))
		}
		b.WriteString("\n[#7c8ca6]FULL PATH[-]\n" + safe(v.Path) + "\n\n[#e8878e]d[-] Review permanent cleanup\n[#7c8ca6]Git-tracked files and symlinks are refused.[-]")
	case DiskEntry:
		heading("PATH", v.Path)
		heading("ALLOCATED", bytesLabel(v.Bytes))
		if u.storage.Total > 0 {
			pct := float64(v.Bytes) * 100 / float64(u.storage.Total)
			fmt.Fprintf(&b, "%s  %.1f%%\n\n", meter(pct, 22, "#b8a0e4"), pct)
		}
		b.WriteString("[#7c8ca6]Allocated blocks, decimal units.\nApplication bundles and immediate home\nfolders are measured recursively.\n\nUse Reclaim to remove generated artifacts.[-]")
		if len(u.storage.Warnings) > 0 {
			b.WriteString("\n\n[#dcb77a]SCAN NOTES[-]\n" + safe(strings.Join(u.storage.Warnings, "\n")))
		}
	}
	u.detail.SetText(b.String()).ScrollToBeginning()
}
func (u *UI) updateDashboard() {
	m := u.metrics
	host, _ := os.Hostname()
	u.header.SetText(fmt.Sprintf(" [#58d6be::b]H A R B O R[-::-] [#7c8ca6]· %s · macOS[-]                                        [#7c8ca6]%s[-]", safe(strings.TrimSuffix(host, ".local")), time.Now().Format("15:04:05")))
	_, _, width, _ := u.cpuPanel.GetInnerRect()
	width = max(10, width-2)
	status := fmt.Sprintf("[#7c8ca6]utilization · %d processes[-]", m.Processes)
	if m.Error != "" {
		status = "[#e8878e]sampler restarting · " + safe(strings.SplitN(m.Error, "\n", 2)[0]) + "[-]"
	}
	if len(u.cpuHistory) == 0 {
		if m.Error == "" {
			status = "[#7c8ca6]Collecting live samples…[-]"
		}
		u.cpuPanel.SetText(" " + status + "\n\n [#26364a]────────────────────────────[-]")
	} else {
		u.cpuPanel.SetText(fmt.Sprintf(" [#58d6be::b]%5.1f%%[-::-]  %s\n%s\n [#7c8ca6]load %s[-]", m.CPU, status, graph(u.cpuHistory, width, 2), safe(m.Load)))
	}
	_, _, memWidth, _ := u.memoryPanel.GetInnerRect()
	memWidth = max(8, memWidth-3)
	memTotal := m.MemoryUsed + m.MemoryFree
	if memTotal > 0 {
		pct := float64(m.MemoryUsed) * 100 / float64(memTotal)
		u.memoryPanel.SetText(fmt.Sprintf(" [#b8a0e4::b]%s[-::-] [#7c8ca6]used[-]\n %s\n [#7c8ca6]free[-]  %s\n [#7c8ca6]total[-] %s", bytesLabel(m.MemoryUsed), meter(pct, memWidth, "#b8a0e4"), bytesLabel(m.MemoryFree), bytesLabel(memTotal)))
	} else {
		u.memoryPanel.SetText(" [#7c8ca6]Collecting…[-]")
	}
	_, _, diskWidth, _ := u.diskPanel.GetInnerRect()
	diskWidth = max(8, diskWidth-3)
	if m.DiskTotal > 0 {
		pct := float64(m.DiskTotal-m.DiskFree) * 100 / float64(m.DiskTotal)
		u.diskPanel.SetText(fmt.Sprintf(" [#dcb77a::b]%s[-::-] [#7c8ca6]used[-]\n %s\n [#7c8ca6]free[-]  %s\n [#7c8ca6]total[-] %s", bytesLabel(m.DiskTotal-m.DiskFree), meter(pct, diskWidth, "#dcb77a"), bytesLabel(m.DiskFree), bytesLabel(m.DiskTotal)))
	} else {
		u.diskPanel.SetText(" [#7c8ca6]Collecting…[-]")
	}
}
func (u *UI) applyMetrics(m Metrics) {
	u.metrics = m
	u.cpuHistory = append(u.cpuHistory, m.CPU)
	if len(u.cpuHistory) > 120 {
		u.cpuHistory = u.cpuHistory[len(u.cpuHistory)-120:]
	}
	u.updateDashboard()
}
func (u *UI) watchMetrics() {
	for u.ctx.Err() == nil {
		err := streamMetrics(u.ctx, u.home, 2, func(m Metrics) { u.queue(func() { u.applyMetrics(m) }) })
		if u.ctx.Err() != nil {
			return
		}
		u.queue(func() { u.metrics.Error = err.Error(); u.updateDashboard() })
		select {
		case <-u.ctx.Done():
		case <-time.After(2 * time.Second):
		}
	}
}
func firstLine(s string) string { return strings.SplitN(s, "\n", 2)[0] }
func (u *UI) flash(kind, message string) {
	u.status, u.statusKind, u.statusAt = message, kind, time.Now()
	u.renderFooter()
}
func (u *UI) renderFooter() {
	line := " [#58d6be]" + u.hint + "[-]"
	if u.status != "" && (u.statusKind == "busy" || time.Since(u.statusAt) < 8*time.Second) {
		color := map[string]string{"busy": "#dcb77a", "ok": "#58d6be", "err": "#e8878e", "info": "#72bee8"}[u.statusKind]
		icon := map[string]string{"busy": "⟳", "ok": "✔", "err": "✖", "info": "›"}[u.statusKind]
		text := u.status
		if u.statusKind == "busy" {
			text += fmt.Sprintf(" · %ds", int(time.Since(u.statusAt).Seconds()))
		}
		line = fmt.Sprintf(" [#10151f:%s:b] %s %s [-:-:-]", color, icon, safe(text))
	}
	u.footer.SetText(line + "\n [#7c8ca6]1–6 tabs  / filter  r refresh  tab focus  ? help  q quit[-]")
}
func setDarkButtons(form *tview.Form) {
	form.SetFieldBackgroundColor(tcell.GetColor("#172230")).SetFieldTextColor(ink).SetLabelColor(teal).SetButtonBackgroundColor(tcell.GetColor("#223a46")).SetButtonTextColor(teal)
}
