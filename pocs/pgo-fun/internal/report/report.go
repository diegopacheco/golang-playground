package report

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"
)

var ErrFormat = errors.New("bad results line")

type Run struct {
	Label    string
	RPS      float64
	P50      time.Duration
	P99      time.Duration
	Requests int
	Errors   int
}

func (r Run) Line() string {
	return fmt.Sprintf("%s\t%.1f\t%d\t%d\t%d\t%d", r.Label, r.RPS, r.P50.Microseconds(), r.P99.Microseconds(), r.Requests, r.Errors)
}

func Parse(in io.Reader) ([]Run, error) {
	var runs []Run
	sc := bufio.NewScanner(in)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) != 6 {
			return nil, fmt.Errorf("%w %d: want 6 fields, got %d", ErrFormat, n, len(f))
		}
		rps, err1 := strconv.ParseFloat(f[1], 64)
		p50, err2 := strconv.ParseInt(f[2], 10, 64)
		p99, err3 := strconv.ParseInt(f[3], 10, 64)
		reqs, err4 := strconv.Atoi(f[4])
		errs, err5 := strconv.Atoi(f[5])
		if err := errors.Join(err1, err2, err3, err4, err5); err != nil {
			return nil, fmt.Errorf("%w %d: %w", ErrFormat, n, err)
		}
		runs = append(runs, Run{Label: f[0], RPS: rps, P50: time.Duration(p50) * time.Microsecond, P99: time.Duration(p99) * time.Microsecond, Requests: reqs, Errors: errs})
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read results: %w", err)
	}
	return runs, nil
}

type Summary struct {
	Label  string
	Rounds int
	RPS    float64
	P50    time.Duration
	P99    time.Duration
	Errors int
}

func Summarize(runs []Run, label string) (Summary, error) {
	var rps, p50, p99 []float64
	s := Summary{Label: label}
	for _, r := range runs {
		if r.Label != label {
			continue
		}
		rps = append(rps, r.RPS)
		p50 = append(p50, float64(r.P50))
		p99 = append(p99, float64(r.P99))
		s.Errors += r.Errors
	}
	if len(rps) == 0 {
		return Summary{}, fmt.Errorf("no runs labeled %q", label)
	}
	s.Rounds = len(rps)
	s.RPS = median(rps)
	s.P50 = time.Duration(median(p50))
	s.P99 = time.Duration(median(p99))
	return s, nil
}

func median(v []float64) float64 {
	s := slices.Clone(v)
	slices.Sort(s)
	m := len(s) / 2
	if len(s)%2 == 1 {
		return s[m]
	}
	return (s[m-1] + s[m]) / 2
}

func Delta(base, cand float64) float64 {
	if base == 0 {
		return 0
	}
	return (cand - base) / base * 100
}

func Write(w io.Writer, runs []Run, base, cand Summary) error {
	var b strings.Builder
	fmt.Fprintf(&b, "%-6s %-9s %12s %10s %10s %8s\n", "round", "build", "req/s", "p50", "p99", "errors")
	round := map[string]int{}
	for _, r := range runs {
		round[r.Label]++
		fmt.Fprintf(&b, "%-6d %-9s %12.1f %10s %10s %8d\n", round[r.Label], r.Label, r.RPS, ms(r.P50), ms(r.P99), r.Errors)
	}
	b.WriteString("\n")
	fmt.Fprintf(&b, "%-14s %14s %14s %9s\n", "median", base.Label, cand.Label, "delta")
	fmt.Fprintf(&b, "%-14s %14.1f %14.1f %+8.1f%%\n", "req/s", base.RPS, cand.RPS, Delta(base.RPS, cand.RPS))
	fmt.Fprintf(&b, "%-14s %14s %14s %+8.1f%%\n", "p50 latency", ms(base.P50), ms(cand.P50), Delta(float64(base.P50), float64(cand.P50)))
	fmt.Fprintf(&b, "%-14s %14s %14s %+8.1f%%\n", "p99 latency", ms(base.P99), ms(cand.P99), Delta(float64(base.P99), float64(cand.P99)))
	fmt.Fprintf(&b, "%-14s %14d %14d\n", "errors", base.Errors, cand.Errors)
	_, err := io.WriteString(w, b.String())
	return err
}

func ms(d time.Duration) string {
	return fmt.Sprintf("%.3fms", float64(d.Microseconds())/1000)
}
