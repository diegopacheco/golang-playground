package main

import (
	"io"
	"math"
	"strings"
	"testing"
	"time"
)

func TestMetricsUseLatestTopSample(t *testing.T) {
	output := "Processes: 100 total, 2 running\nCPU usage: 20.0% user, 10.0% sys, 70.0% idle\nPhysMem: 8G used (1G wired), 24G unused.\nLoad Avg: 1.0, 2.0, 3.0\nProcesses: 120 total, 3 running\nCPU usage: 5.0% user, 2.5% sys, 92.5% idle\nPhysMem: 9G used (1G wired), 23G unused.\n"
	m := parseTop(output)
	if math.Abs(m.CPU-7.5) > 0.001 || m.Processes != 120 || m.MemoryUsed != 9*1024*1024*1024 || m.MemoryFree != 23*1024*1024*1024 || m.Load != "1.0, 2.0, 3.0" {
		t.Fatalf("%+v", m)
	}
}
func TestStreamEmitsCompletedSamplesLiveAndSkipsBootAverage(t *testing.T) {
	sample := func(idle string) string {
		return "Processes: 10 total\nCPU usage: 1.0% user, 1.0% sys, " + idle + "% idle\nPhysMem: 2G used (1G wired), 6G unused.\nDisks: 1/1G read, 1/1G written.\n"
	}
	r, w := io.Pipe()
	got := make(chan Metrics, 4)
	done := make(chan error, 1)
	go func() { done <- readSamples(r, t.TempDir(), func(m Metrics) { got <- m }) }()
	w.Write([]byte(sample("10.0") + sample("75.0") + "Processes: 11 total\n"))
	select {
	case m := <-got:
		if math.Abs(m.CPU-25) > 0.001 || m.DiskTotal == 0 {
			t.Fatalf("emitted boot average or missing disk: %+v", m)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("completed sample waited for the next interval")
	}
	w.Close()
	if err := <-done; err != nil || len(got) != 0 {
		t.Fatalf("partial sample emitted or read failed: %v %d", err, len(got))
	}
}
func TestGraphDoesNotInventHistory(t *testing.T) {
	text := graph([]float64{100}, 6, 2)
	if strings.Count(text, "█") != 2 || strings.Count(text, "╌") != 10 {
		t.Fatal(text)
	}
	if !strings.HasPrefix(meter(-2, 8, "green"), "[green][#29364a]") {
		t.Fatal("negative value filled meter")
	}
}
