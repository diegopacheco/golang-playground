package runner

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func newRunner(t *testing.T, timeout time.Duration, max int) Runner {
	t.Helper()
	bin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain not found")
	}
	return Runner{GoBin: bin, Timeout: timeout, MaxOutput: max}
}

func TestRun(t *testing.T) {
	r := newRunner(t, 20*time.Second, 1024)
	tests := []struct {
		name       string
		code       string
		wantOut    string
		wantErr    string
		wantExit   int
		wantTimout bool
	}{
		{"valid program prints to stdout so the user sees its output", "package main\nimport \"fmt\"\nfunc main(){ fmt.Println(\"hi\") }\n", "hi\n", "", 0, false},
		{"compile errors reach the user with file and line", "package main\nfunc main(){ x := 1 }\n", "", "main.go:2", 1, false},
		{"the program's own exit code is reported so failures are visible", "package main\nimport \"os\"\nfunc main(){ os.Exit(3) }\n", "", "", 3, false},
		{"panics show the panic message", "package main\nfunc main(){ panic(\"boom\") }\n", "", "panic: boom", 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := r.Run(context.Background(), tt.code)
			if err != nil {
				t.Fatalf("Run error: %v", err)
			}
			if res.Stdout != tt.wantOut {
				t.Errorf("stdout = %q, want %q", res.Stdout, tt.wantOut)
			}
			if !strings.Contains(res.Stderr, tt.wantErr) {
				t.Errorf("stderr = %q, want it to contain %q", res.Stderr, tt.wantErr)
			}
			if res.ExitCode != tt.wantExit {
				t.Errorf("exit = %d, want %d", res.ExitCode, tt.wantExit)
			}
			if strings.Contains(res.Stderr, "goplay-") {
				t.Errorf("stderr leaks the temp dir: %q", res.Stderr)
			}
		})
	}
}

func TestRunKillsInfiniteLoopsSoTheServerStaysResponsive(t *testing.T) {
	r := newRunner(t, 3*time.Second, 1024)
	start := time.Now()
	res, err := r.Run(context.Background(), "package main\nfunc main(){ for {} }\n")
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if !res.TimedOut {
		t.Errorf("TimedOut = false, want true")
	}
	if time.Since(start) > 8*time.Second {
		t.Errorf("run took %s, the timeout did not kill the program", time.Since(start))
	}
}

func TestRunCapsOutputSoAFloodCannotExhaustMemory(t *testing.T) {
	r := newRunner(t, 20*time.Second, 100)
	res, err := r.Run(context.Background(), "package main\nimport \"fmt\"\nfunc main(){ for i := 0; i < 10000; i++ { fmt.Println(i) } }\n")
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if !strings.HasSuffix(res.Stdout, "[output truncated]\n") {
		t.Errorf("stdout not marked truncated: %q", res.Stdout)
	}
	if len(res.Stdout) > 200 {
		t.Errorf("stdout len = %d, want capped near 100", len(res.Stdout))
	}
}

func TestRunRejectsEmptyCode(t *testing.T) {
	r := Runner{GoBin: "go", Timeout: time.Second, MaxOutput: 10}
	if _, err := r.Run(context.Background(), "  \n "); !errors.Is(err, ErrEmptyCode) {
		t.Errorf("err = %v, want ErrEmptyCode", err)
	}
}
