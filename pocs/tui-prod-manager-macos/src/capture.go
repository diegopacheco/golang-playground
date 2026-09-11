package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

type CaptureState struct {
	Processes  []Process
	Agents     []Process
	Containers []Container
	Images     []Image
	Projects   []Project
	Storage    Storage
	Metrics    Metrics
	History    []float64
	Volumes    []Volume
}

func captureScreens(root, dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	u := newUI(root)
	defer u.cancel()
	ctx, cancel := context.WithTimeout(u.ctx, 5*time.Minute)
	defer cancel()
	var err error
	statePath := os.Getenv("HARBOR_CAPTURE_STATE")
	var state CaptureState
	data, readErr := os.ReadFile(statePath)
	if statePath != "" && readErr == nil {
		if err = json.Unmarshal(data, &state); err != nil {
			return err
		}
		u.procs = state.Processes
		u.agents = state.Agents
		u.containers = state.Containers
		u.images = state.Images
		u.projects = state.Projects
		u.storage = state.Storage
		u.metrics = state.Metrics
		u.cpuHistory = state.History
		u.volumes = state.Volumes
	} else {
		u.procs, err = processes(ctx, run)
		if err != nil {
			return err
		}
		for _, p := range u.procs {
			if agentKind(p) != "" || isGoProcess(p) {
				u.agents = append(u.agents, p)
			}
		}
		u.containers, u.images, u.volumes, err = podmanData(ctx, run)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		fmt.Fprintln(os.Stderr, "Capturing reclaim inventory")
		u.projects, _, err = scanProjects(ctx, root, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Reclaim capture:", err)
		}
		fmt.Fprintln(os.Stderr, "Capturing storage inventory")
		u.storage = scanStorage(ctx, u.home, nil)

		sampleCtx, stopSampling := context.WithCancel(ctx)
		sampleErr := streamMetrics(sampleCtx, u.home, 1, func(m Metrics) {
			u.metrics = m
			u.cpuHistory = append(u.cpuHistory, m.CPU)
			if len(u.cpuHistory) == 12 {
				stopSampling()
			}
		})
		if sampleCtx.Err() == nil {
			fmt.Fprintln(os.Stderr, "Metrics capture:", sampleErr)
		}
		stopSampling()
		if statePath != "" {
			state = CaptureState{u.procs, u.agents, u.containers, u.images, u.projects, u.storage, u.metrics, u.cpuHistory, u.volumes}
			data, err = json.Marshal(state)
			if err != nil {
				return err
			}
			if err = os.MkdirAll(filepath.Dir(statePath), 0700); err != nil {
				return err
			}
			if err = os.WriteFile(statePath, data, 0600); err != nil {
				return err
			}
		}
	}
	models, modelErr := modelOptions(ctx, run, u.home, u.config.Provider)
	u.setModels(u.config.Provider, models, modelErr)
	screen := tcell.NewSimulationScreen("UTF-8")
	if err = screen.Init(); err != nil {
		return err
	}
	defer screen.Fini()
	screen.SetSize(144, 46)
	for tab := range tabNames {
		fmt.Fprintln(os.Stderr, "Rendering", tabNames[tab])
		u.loaded[tab] = true
		u.switchTab(tab)
		u.previewID = ""
		u.preview()
		u.pages.SetRect(0, 0, 144, 46)
		screen.Clear()
		u.pages.Draw(screen)
		u.updateDashboard()
		u.pages.Draw(screen)
		screen.Show()
		var b strings.Builder
		b.WriteString("<!doctype html><html><meta charset=\"utf-8\"><title>Harbor — " + tabNames[tab] + "</title><style>body{margin:0;padding:16px;background:#10151f}pre{font:14px/20px 'JetBrains Mono','Menlo',monospace;margin:0;box-shadow:0 16px 48px #25354818;overflow:hidden;width:max-content}span{white-space:pre;display:inline-block;height:20px;line-height:20px;vertical-align:top}</style><body><pre>")
		for y := 0; y < 46; y++ {
			for x := 0; x < 144; x++ {
				r, comb, style, width := screen.GetContent(x, y)
				if width == 0 {
					width = 1
					r = ' '
				}
				if r == 0 {
					r = ' '
				}
				fg, bg, attr := style.Decompose()
				fr, fgc, fb := fg.RGB()
				br, bgc, bb := bg.RGB()
				if fr < 0 {
					fr, fgc, fb = 215, 224, 237
				}
				if br < 0 {
					br, bgc, bb = 16, 21, 31
				}
				weight := "normal"
				if attr&tcell.AttrBold != 0 {
					weight = "bold"
				}
				fmt.Fprintf(&b, "<span style=\"color:rgb(%d,%d,%d);background:rgb(%d,%d,%d);font-weight:%s;width:%dch\">%s</span>", fr, fgc, fb, br, bgc, bb, weight, width, html.EscapeString(string(append([]rune{r}, comb...))))
				if width > 1 {
					x += width - 1
				}
			}
			b.WriteByte('\n')
		}
		b.WriteString("</pre></body></html>")
		if err = os.WriteFile(filepath.Join(dir, fmt.Sprintf("%02d-%s.html", tab+1, strings.ToLower(tabNames[tab]))), []byte(b.String()), 0644); err != nil {
			return err
		}
	}
	return nil
}
