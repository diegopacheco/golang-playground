package main

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var sortNames = []string{"CPU", "memory", "age"}
var condition = regexp.MustCompile(`^(cpu|mem|age)(>=|<=|>|<)(.+)$`)
var percentUnits = map[string]float64{"": 1, "%": 1}
var byteUnits = map[string]float64{"": 1, "b": 1, "k": 1e3, "kb": 1e3, "m": 1e6, "mb": 1e6, "g": 1e9, "gb": 1e9, "t": 1e12, "tb": 1e12}
var timeUnits = map[string]float64{"": 1, "s": 1, "m": 60, "h": 3600, "d": 86400}

func elapsedSeconds(s string) int64 {
	var days int64
	if d, rest, ok := strings.Cut(s, "-"); ok {
		days, _ = strconv.ParseInt(d, 10, 64)
		s = rest
	}
	var total int64
	for _, part := range strings.Split(s, ":") {
		n, _ := strconv.ParseInt(part, 10, 64)
		total = total*60 + n
	}
	return days*86400 + total
}
func scaled(s string, units map[string]float64) (float64, bool) {
	i := strings.LastIndexAny(s, "0123456789.") + 1
	n, err := strconv.ParseFloat(s[:i], 64)
	factor, ok := units[s[i:]]
	return n * factor, err == nil && ok
}
func processMatches(p Process, token string) (matched, handled bool) {
	m := condition.FindStringSubmatch(token)
	if m == nil {
		return false, false
	}
	var value, limit float64
	var ok bool
	switch m[1] {
	case "cpu":
		value, _ = strconv.ParseFloat(p.CPU, 64)
		limit, ok = scaled(m[3], percentUnits)
	case "mem":
		value = float64(p.Memory)
		limit, ok = scaled(m[3], byteUnits)
	case "age":
		value = float64(elapsedSeconds(p.Elapsed))
		limit, ok = scaled(m[3], timeUnits)
	}
	if !ok {
		return false, true
	}
	switch m[2] {
	case ">":
		return value > limit, true
	case "<":
		return value < limit, true
	case ">=":
		return value >= limit, true
	}
	return value <= limit, true
}
func matchesQuery(query, search string, item any) bool {
	search = strings.ToLower(search)
	for _, token := range strings.Fields(strings.ToLower(query)) {
		if p, ok := item.(Process); ok {
			if matched, handled := processMatches(p, token); handled {
				if !matched {
					return false
				}
				continue
			}
		}
		if !strings.Contains(search, token) {
			return false
		}
	}
	return true
}
func sortProcesses(ps []Process, by int) []Process {
	sorted := append([]Process{}, ps...)
	key := func(p Process) float64 {
		switch by {
		case 1:
			return float64(p.Memory)
		case 2:
			return float64(elapsedSeconds(p.Elapsed))
		}
		cpu, _ := strconv.ParseFloat(p.CPU, 64)
		return cpu
	}
	sort.SliceStable(sorted, func(i, j int) bool { return key(sorted[i]) > key(sorted[j]) })
	return sorted
}
