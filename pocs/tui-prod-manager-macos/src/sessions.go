package main

import (
	"bufio"
	"context"
	"debug/buildinfo"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var goBinaries sync.Map

func isGoProcess(p Process) bool {
	if agentKind(p) == "Go" {
		return true
	}
	if !filepath.IsAbs(p.Executable) {
		return false
	}
	if v, ok := goBinaries.Load(p.Executable); ok {
		return v.(bool)
	}
	_, err := buildinfo.ReadFile(p.Executable)
	goBinaries.Store(p.Executable, err == nil)
	return err == nil
}

type LogSource struct {
	Path, Label string
	Modified    time.Time
	Linked      bool
}

func isLogFile(name string) bool {
	return strings.HasSuffix(name, ".log") || strings.HasSuffix(name, ".jsonl") || strings.HasSuffix(name, ".out")
}
func sessionSources(ctx context.Context, runner Runner, home string, p Process, all bool) ([]LogSource, error) {
	result := []LogSource{}
	seen := map[string]bool{}
	add := func(path, label string, linked bool) {
		if seen[path] {
			return
		}
		if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
			seen[path] = true
			result = append(result, LogSource{path, label, info.ModTime(), linked})
		}
	}
	kind := agentKind(p)
	if kind == "Claude" {
		dir := os.Getenv("CLAUDE_CONFIG_DIR")
		if dir == "" {
			dir = filepath.Join(home, ".claude")
		}
		data, err := os.ReadFile(filepath.Join(dir, "sessions", strconv.Itoa(p.PID)+".json"))
		if err == nil {
			var v struct {
				PID       int    `json:"pid"`
				SessionID string `json:"sessionId"`
				CWD       string `json:"cwd"`
			}
			if json.Unmarshal(data, &v) == nil && v.PID == p.PID && v.SessionID != "" {
				encoded := strings.Map(func(r rune) rune {
					if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
						return r
					}
					return '-'
				}, v.CWD)
				add(filepath.Join(dir, "projects", encoded, v.SessionID+".jsonl"), "PID-linked Claude transcript", true)
			}
		}
	}
	output, err := runner(ctx, "lsof", "-a", "-p", strconv.Itoa(p.PID), "-Fftn")
	cwd, fd, fileType := "", "", ""
	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}
		value := line[1:]
		switch line[0] {
		case 'f':
			fd = value
		case 't':
			fileType = value
		case 'n':
			switch {
			case fd == "cwd":
				cwd = value
			case fileType != "REG":
			case fd == "1" || fd == "2":
				add(value, "Process output", true)
			case isLogFile(value):
				add(value, "Open log file", true)
			}
		}
	}
	if cwd != "" && (all || len(result) == 0) {
		entries, _ := os.ReadDir(cwd)
		for _, e := range entries {
			if !e.IsDir() && isLogFile(e.Name()) {
				add(filepath.Join(cwd, e.Name()), "Log in working directory", false)
			}
		}
	}
	if (kind == "Codex" || kind == "Claude") && (all || len(result) == 0) {
		root := filepath.Join(home, ".codex", "sessions")
		if configured := os.Getenv("CODEX_HOME"); configured != "" {
			root = filepath.Join(configured, "sessions")
		}
		if kind == "Claude" {
			root = filepath.Join(home, ".claude", "projects")
			if configured := os.Getenv("CLAUDE_CONFIG_DIR"); configured != "" {
				root = filepath.Join(configured, "projects")
			}
		}
		files := []LogSource{}
		walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if e != nil {
				return nil
			}
			if !d.IsDir() && strings.HasSuffix(path, ".jsonl") {
				if s, e := d.Info(); e == nil {
					files = append(files, LogSource{Path: path, Modified: s.ModTime()})
				}
			}
			return nil
		})
		if walkErr != nil {
			return nil, walkErr
		}
		sort.Slice(files, func(i, j int) bool { return files[i].Modified.After(files[j].Modified) })
		for _, f := range files[:min(300, len(files))] {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			file, e := os.Open(f.Path)
			if e != nil {
				continue
			}
			reader := bufio.NewScanner(io.LimitReader(file, 128*1024))
			match := false
			for i := 0; i < 8 && reader.Scan(); i++ {
				var v struct {
					CWD     string `json:"cwd"`
					Payload struct {
						CWD string `json:"cwd"`
					} `json:"payload"`
				}
				if json.Unmarshal(reader.Bytes(), &v) == nil && cwd != "" && (v.CWD == cwd || v.Payload.CWD == cwd) {
					match = true
					break
				}
			}
			file.Close()
			if match {
				add(f.Path, "Same working directory", false)
			}
			if len(result) >= 12 {
				break
			}
		}
		if all || len(result) == 0 {
			for _, f := range files[:min(12, len(files))] {
				add(f.Path, "Recent "+kind+" transcript", false)
			}
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("PID %d has no transcript, open log file, stdout/stderr file, or *.log in its working directory. Terminal output cannot be read from another PTY; redirect it to a .log file to follow it here. Inspection: %v", p.PID, err)
	}
	return result, nil
}
func contentText(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []any:
		parts := []string{}
		for _, item := range x {
			if text := contentText(item); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "\n")
	case map[string]any:
		if text, ok := x["text"].(string); ok {
			return text
		}
	}
	return ""
}
func transcriptLine(line string) string {
	var v map[string]any
	if json.Unmarshal([]byte(line), &v) != nil {
		return line
	}
	if msg, ok := v["message"].(map[string]any); ok {
		if text := contentText(msg["content"]); text != "" {
			return fmt.Sprintf("%v  %s", msg["role"], text)
		}
	}
	if p, ok := v["payload"].(map[string]any); ok {
		if p["type"] == "message" {
			return fmt.Sprintf("%v  %s", p["role"], contentText(p["content"]))
		}
		if v["type"] == "event_msg" && (p["type"] == "agent_message" || p["type"] == "user_message") {
			return fmt.Sprintf("%v  %s", p["type"], contentText(p["message"]))
		}
	}
	return ""
}
func tailLog(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	s, err := f.Stat()
	if err != nil {
		return "", err
	}
	if !s.Mode().IsRegular() {
		return "", fmt.Errorf("log must be a regular file")
	}
	offset := max(int64(0), s.Size()-256*1024)
	if _, err = f.Seek(offset, io.SeekStart); err != nil {
		return "", err
	}
	data, err := io.ReadAll(io.LimitReader(f, 256*1024))
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	if offset > 0 && len(lines) > 0 {
		lines = lines[1:]
	}
	if strings.HasSuffix(path, ".jsonl") && len(lines) > 0 {
		lines = lines[:len(lines)-1]
	}
	var b strings.Builder
	for _, line := range lines {
		text := line
		if strings.HasSuffix(path, ".jsonl") {
			text = transcriptLine(line)
		}
		if strings.TrimSpace(text) != "" {
			b.WriteString(text)
			b.WriteString("\n\n")
		}
	}
	if b.Len() == 0 {
		return "Waiting for readable messages…", nil
	}
	return b.String(), nil
}
