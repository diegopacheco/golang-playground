package report

import (
	"fmt"
	"math"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/guptarohit/asciigraph"

	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/sim"
)

const (
	maxRows = 30
	bold    = "\x1b[1m"
	reset   = "\x1b[0m"
)

var palette = []asciigraph.AnsiColor{
	asciigraph.HotPink,
	asciigraph.DodgerBlue,
	asciigraph.Gold,
	asciigraph.LimeGreen,
	asciigraph.DarkOrange,
	asciigraph.Cyan,
}

type Report struct {
	Spec    sim.Spec
	Limits  sim.Limits
	Offered []float64
	Results []sim.Result
	Color   bool
}

func (r Report) Header() string {
	return fmt.Sprintf("%s\nrate %d/s (one every %v), burst or queue %d, window %v limit %d, tick %v, %d ticks\nload: %s",
		r.style("RATE LIMITER LAB"),
		int(time.Second/r.Limits.Every), r.Limits.Every, r.Limits.Burst, r.Limits.Window, r.Limits.Limit, r.Spec.Tick, len(r.Offered),
		phases(r.Spec))
}

func (r Report) OfferedPlot() string {
	return r.plot(r.Offered, "offered requests per tick", asciigraph.LightGray, maxOf(r.Offered))
}

func (r Report) Plots() string {
	upper := 0.0
	for _, res := range r.Results {
		upper = max(upper, maxOf(res.PerTick))
	}
	var parts []string
	for i, res := range r.Results {
		parts = append(parts, r.plot(res.PerTick, res.Name+" · allowed per tick", palette[i%len(palette)], upper))
	}
	return strings.Join(parts, "\n\n")
}

func (r Report) Summary() string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ALGORITHM\tALLOWED\tREJECTED\tPEAK/TICK\tPEAK/WINDOW\tMAX DELAY\tSTATE")
	for _, res := range r.Results {
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%v\t%s\n", res.Name, res.Allowed, res.Rejected, res.PeakTick, res.PeakWindow, res.MaxDelay, res.State)
	}
	w.Flush()
	lines := strings.Split(strings.TrimRight(b.String(), "\n"), "\n")
	lines[0] = r.style(lines[0])
	return strings.Join(lines, "\n") + fmt.Sprintf("\n\nPEAK/WINDOW is the most requests let through in any %v span; the limit is %d.", r.Spec.Window, r.Limits.Limit)
}

func (r Report) String() string {
	return strings.Join([]string{r.Header(), r.OfferedPlot(), r.Plots(), r.Summary()}, "\n\n") + "\n"
}

func axis(peak float64) (float64, int) {
	rows := int(math.Ceil(peak))
	if rows <= maxRows {
		return float64(rows), rows
	}
	step := math.Ceil(peak / maxRows)
	return step * maxRows, maxRows
}

func (r Report) plot(series []float64, caption string, color asciigraph.AnsiColor, peak float64) string {
	upper, rows := axis(peak)
	opts := []asciigraph.Option{
		asciigraph.Height(rows),
		asciigraph.LowerBound(0),
		asciigraph.UpperBound(upper),
		asciigraph.Precision(0),
		asciigraph.Caption(caption),
	}
	if r.Color {
		opts = append(opts, asciigraph.SeriesColors(color), asciigraph.CaptionColor(color), asciigraph.AxisColor(asciigraph.DimGray), asciigraph.LabelColor(asciigraph.LightGray))
	}
	return asciigraph.Plot(series, opts...)
}

func (r Report) style(s string) string {
	if !r.Color {
		return s
	}
	return bold + s + reset
}

func phases(spec sim.Spec) string {
	var parts []string
	for _, p := range spec.Pattern {
		parts = append(parts, fmt.Sprintf("%s %dx%d", p.Name, p.Ticks, p.PerTick))
	}
	return strings.Join(parts, " → ")
}

func maxOf(series []float64) float64 {
	m := 1.0
	for _, v := range series {
		m = max(m, v)
	}
	return m
}
