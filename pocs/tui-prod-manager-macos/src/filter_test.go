package main

import (
	"reflect"
	"testing"
)

func TestProcessConditionsFilterByCPUMemoryAndAge(t *testing.T) {
	ps := []Process{
		{PID: 1, CPU: "12.5", Memory: 2e9, Elapsed: "2-00:00:00"},
		{PID: 2, CPU: "0.5", Memory: 100e6, Elapsed: "05:00"},
		{PID: 3, CPU: "50", Memory: 600e6, Elapsed: "01:00:00"},
	}
	pids := func(query string) []int {
		var result []int
		for _, p := range ps {
			if matchesQuery(query, "chrome", p) {
				result = append(result, p.PID)
			}
		}
		return result
	}
	for query, want := range map[string][]int{
		"cpu>5 mem>1gb": {1},
		"cpu>=50":       {3},
		"age<1h":        {2},
		"age>=1h":       {1, 3},
		"mem<500MB":     {2},
		"cpu>5 chrome":  {1, 3},
		"cpu>5 firefox": nil,
		"cpu>abc":       nil,
	} {
		if got := pids(query); !reflect.DeepEqual(got, want) {
			t.Errorf("%q matched %v, want %v", query, got, want)
		}
	}
}
func TestElapsedSecondsReadsPsFormat(t *testing.T) {
	for in, want := range map[string]int64{"2-03:04:05": 183845, "01:02:03": 3723, "04:05": 245} {
		if got := elapsedSeconds(in); got != want {
			t.Errorf("%s = %d, want %d", in, got, want)
		}
	}
}
