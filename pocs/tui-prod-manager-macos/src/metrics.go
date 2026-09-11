package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

type Metrics struct {
	CPU                                         float64
	MemoryUsed, MemoryFree, DiskTotal, DiskFree int64
	Processes                                   int
	Load                                        string
	Error                                       string
}

var cpuPattern = regexp.MustCompile(`CPU usage:.*?([\d.]+)% idle`)
var memoryPattern = regexp.MustCompile(`PhysMem:\s+(\S+) used.*?\s(\S+) unused`)
var processPattern = regexp.MustCompile(`Processes:\s+(\d+) total`)

func topBytes(s string) int64 {
	s = strings.TrimRight(s, ".,")
	if s == "" {
		return 0
	}
	unit := s[len(s)-1]
	factor := float64(1)
	switch unit {
	case 'K':
		factor = 1024
	case 'M':
		factor = 1024 * 1024
	case 'G':
		factor = 1024 * 1024 * 1024
	case 'T':
		factor = 1024 * 1024 * 1024 * 1024
	}
	if factor > 1 || unit == 'B' {
		s = s[:len(s)-1]
	}
	n, _ := strconv.ParseFloat(s, 64)
	return int64(n * factor)
}
func parseTop(output string) Metrics {
	m := Metrics{}
	for _, line := range strings.Split(output, "\n") {
		if match := cpuPattern.FindStringSubmatch(line); len(match) > 1 {
			idle, _ := strconv.ParseFloat(match[1], 64)
			m.CPU = math.Max(0, math.Min(100, 100-idle))
		}
		if match := memoryPattern.FindStringSubmatch(line); len(match) > 2 {
			m.MemoryUsed = topBytes(match[1])
			m.MemoryFree = topBytes(match[2])
		}
		if match := processPattern.FindStringSubmatch(line); len(match) > 1 {
			m.Processes, _ = strconv.Atoi(match[1])
		}
		if strings.HasPrefix(line, "Load Avg:") {
			m.Load = strings.TrimSpace(strings.TrimPrefix(line, "Load Avg:"))
		}
	}
	return m
}
func readSamples(r io.Reader, home string, emit func(Metrics)) error {
	scanner := bufio.NewScanner(r)
	var sample strings.Builder
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		sample.WriteString(line)
		sample.WriteByte('\n')
		if !strings.HasPrefix(line, "Disks:") {
			continue
		}
		if !first {
			m := parseTop(sample.String())
			var stat syscall.Statfs_t
			if err := syscall.Statfs(home, &stat); err == nil {
				m.DiskTotal = int64(stat.Blocks) * int64(stat.Bsize)
				m.DiskFree = int64(stat.Bavail) * int64(stat.Bsize)
			}
			emit(m)
		}
		first = false
		sample.Reset()
	}
	return scanner.Err()
}
func streamMetrics(ctx context.Context, home string, seconds int, emit func(Metrics)) error {
	cmd := exec.CommandContext(ctx, "top", "-l", "0", "-s", strconv.Itoa(seconds), "-n", "0")
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}
	readErr := readSamples(out, home, emit)
	if err = cmd.Wait(); err != nil {
		return err
	}
	if readErr != nil {
		return readErr
	}
	return errors.New("top sampler exited")
}
func meter(pct float64, width int, color string) string {
	width = max(1, width)
	filled := int(math.Round(math.Max(0, math.Min(100, pct)) * float64(width) / 100))
	return "[" + color + "]" + strings.Repeat("━", filled) + "[#29364a]" + strings.Repeat("━", width-filled) + "[-]"
}
func graph(history []float64, width, height int) string {
	width = max(1, width)
	height = max(1, height)
	history = history[max(0, len(history)-width):]
	var b strings.Builder
	for row := height - 1; row >= 0; row-- {
		b.WriteString(" ")
		for col := 0; col < width; col++ {
			index := col - (width - len(history))
			level := 0
			if index >= 0 {
				level = int(math.Round(history[index]*float64(height*8)/100)) - row*8
			}
			if level <= 0 {
				b.WriteString("[#26364a]╌")
			} else {
				runes := []rune("▁▂▃▄▅▆▇█")
				color := "#58d6be"
				if row > 0 {
					color = "#72bee8"
				}
				b.WriteString(fmt.Sprintf("[%s]%c", color, runes[min(8, level)-1]))
			}
		}
		b.WriteString("[-]")
		if row > 0 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
