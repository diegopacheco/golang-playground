package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestProcessesPreserveFullPathsAndSortCPU(t *testing.T) {
	runner := func(context.Context, string, ...string) (string, error) {
		return " 12 1 diego 0.2 100 01:20 /Applications/Long App.app/Contents/MacOS/Long App\n 13 1 diego 5.0 200 00:12 /usr/local/bin/claude\n", nil
	}
	ps, err := processes(context.Background(), runner)
	if err != nil || len(ps) != 2 {
		t.Fatalf("%v %v", ps, err)
	}
	if ps[0].PID != 13 || ps[1].Executable != "/Applications/Long App.app/Contents/MacOS/Long App" || ps[1].Memory != 102400 {
		t.Fatal(ps)
	}
}
func TestProcessCollectionFailureIsVisible(t *testing.T) {
	_, err := processes(context.Background(), func(context.Context, string, ...string) (string, error) { return "", errors.New("permission denied") })
	if err == nil {
		t.Fatal("collection failure was hidden")
	}
}
func TestSignalsRefuseProtectedAndReusedPIDs(t *testing.T) {
	for _, pid := range []int{0, 1, os.Getpid(), os.Getppid()} {
		if err := signalProcess(context.Background(), nil, Process{PID: pid}, "", syscall.SIGTERM); err == nil {
			t.Fatalf("accepted PID %d", pid)
		}
	}
	runner := func(context.Context, string, ...string) (string, error) { return "new process", nil }
	if err := signalProcess(context.Background(), runner, Process{PID: 999999}, "old process", syscall.SIGTERM); err == nil {
		t.Fatal("reused PID accepted")
	}
}
func TestSignalOwnedChild(t *testing.T) {
	cmd := exec.Command("/bin/cat")
	input, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()
	runner := func(context.Context, string, ...string) (string, error) { return "owned child", nil }
	if err = signalProcess(context.Background(), runner, Process{PID: cmd.Process.Pid}, "owned child", syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if err = cmd.Wait(); err == nil {
		t.Fatal("child did not receive signal")
	}
}
func TestPodmanParsingAndActions(t *testing.T) {
	runner := func(_ context.Context, name string, args ...string) (string, error) {
		if name != "podman" {
			t.Fatal(name)
		}
		if args[0] == "ps" {
			return `[{"Id":"abc123","Names":["db"],"State":"running","Ports":[{"container_port":22,"host_port":2222,"protocol":"tcp"}]}]`, nil
		}
		if args[0] == "volume" {
			return `[{"Name":"pgdata","Driver":"local","MountCount":0}]`, nil
		}
		return `[{"Id":"def456","RepoTags":null,"Names":["localhost/app:v1"],"Size":123456,"Containers":2}]`, nil
	}
	cs, ims, vols, err := podmanData(context.Background(), runner)
	if err != nil || len(cs) != 1 || cs[0].ID != "abc123" || ims[0].Size != 123456 || len(ims[0].tags()) != 1 || ims[0].tags()[0] != "localhost/app:v1" || ims[0].Containers != 2 {
		t.Fatalf("%+v %+v %v", cs, ims, err)
	}
	if len(vols) != 1 || vols[0].Name != "pgdata" {
		t.Fatalf("volumes not listed: %+v", vols)
	}
	args, err := sshArgs(cs[0], "dev")
	if err != nil || !reflect.DeepEqual(args, []string{"-p", "2222", "-l", "dev", "--", "127.0.0.1"}) {
		t.Fatalf("%v %v", args, err)
	}
	capture := func(id, action string) []string {
		var got []string
		if _, err := containerAction(context.Background(), func(_ context.Context, _ string, args ...string) (string, error) { got = args; return "", nil }, id, action); err != nil {
			t.Fatal(err)
		}
		return got
	}
	for _, action := range []string{"start", "stop", "restart", "kill"} {
		got := capture("abc123", action)
		if got[0] != action || got[len(got)-1] != "abc123" || strings.Contains(strings.Join(got, " "), "--force") {
			t.Fatalf("lifecycle actions must never escalate to force: %v", got)
		}
	}
	if got := capture("def456", "rmi"); !reflect.DeepEqual(got, []string{"rmi", "--force", "def456"}) {
		t.Fatalf("exited containers must not block image deletion: %v", got)
	}
	if got := capture("pgdata", "volume-rm"); !reflect.DeepEqual(got, []string{"volume", "rm", "--force", "pgdata"}) {
		t.Fatalf("volume delete must force: %v", got)
	}
	if _, err = containerAction(context.Background(), runner, "--all", "kill"); err == nil {
		t.Fatal("option injection accepted")
	}
	if _, err = sshArgs(Container{}, "dev"); err == nil {
		t.Fatal("missing SSH mapping accepted")
	}
}
func TestCommandDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err := run(ctx, "/bin/sh", "-c", "while :; do :; done")
	if err == nil || ctx.Err() == nil {
		t.Fatal("deadline did not stop process")
	}
}
func TestStorageLabels(t *testing.T) {
	for n, want := range map[int64]string{999: "999 B", 1000: "1.0 kB", 1234000: "1.2 MB", 2000000000: "2.0 GB"} {
		if got := bytesLabel(n); got != want {
			t.Fatalf("%d: %s", n, got)
		}
	}
}
func TestAIUsesLocalSDKWithLiteralPromptAndRestrictions(t *testing.T) {
	for _, provider := range []string{"claude", "codex", "ollama"} {
		t.Run(provider, func(t *testing.T) {
			var command []string
			runner := func(_ context.Context, name string, args ...string) (string, error) {
				command = append([]string{name}, args...)
				return "A PID identifies a running process.", nil
			}
			text, err := explain(context.Background(), runner, AIConfig{provider, "test-model"}, "What is a PID? $(touch unwanted)", "PID 42")
			if err != nil || text == "" {
				t.Fatal(err)
			}
			if command[0] != provider || !strings.Contains(strings.Join(command, "\n"), "$(touch unwanted)") {
				t.Fatal(command)
			}
			if provider == "claude" && !strings.Contains(strings.Join(command, " "), "--tools  --strict-mcp-config") {
				t.Fatal(command)
			}
			if provider == "claude" && command[len(command)-2] != "--" {
				t.Fatal("prompt must follow the variadic MCP option terminator")
			}
			if provider == "codex" && !strings.Contains(strings.Join(command, " "), "--sandbox read-only") {
				t.Fatal(command)
			}
		})
	}
}
func TestAIConfigurationRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	want := AIConfig{"ollama", "local-model"}
	if err := saveConfig(want); err != nil {
		t.Fatal(err)
	}
	got, err := loadConfig()
	if err != nil || got != want {
		t.Fatalf("%v %v", got, err)
	}
	info, _ := os.Stat(configPath())
	if info.Mode().Perm() != 0600 {
		t.Fatal(info.Mode())
	}
	if err = saveConfig(AIConfig{"unknown", "x"}); err == nil {
		t.Fatal("invalid provider accepted")
	}
}
func TestLogParsingAndPartialRecords(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "session.jsonl")
	data := `{"message":{"role":"assistant","content":[{"type":"text","text":"Claude response"}]}}
{"type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"Codex response"}]}}
{"type":"event_msg","payload":{"type":"agent_message","message":"Still working"}}
{"unfinished":`
	if err := os.WriteFile(p, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	text, err := tailLog(p)
	if err != nil || !strings.Contains(text, "Claude response") || !strings.Contains(text, "Codex response") || !strings.Contains(text, "Still working") || strings.Contains(text, "unfinished") {
		t.Fatalf("%s %v", text, err)
	}
	f, _ := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0600)
	f.WriteString("true}\n")
	f.WriteString("{\"message\":{\"role\":\"user\",\"content\":\"new message\"}}\n")
	f.Close()
	text, err = tailLog(p)
	if err != nil || !strings.Contains(text, "new message") {
		t.Fatal(text, err)
	}
	if _, err = tailLog(dir); err == nil {
		t.Fatal("directory accepted as log")
	}
}
