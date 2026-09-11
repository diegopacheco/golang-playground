package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Runner func(context.Context, string, ...string) (string, error)

type limitedBuffer struct {
	bytes.Buffer
	limit int
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	if b.Len() < b.limit {
		b.Buffer.Write(p[:min(len(p), b.limit-b.Len())])
	}
	return n, nil
}
func run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = time.Second
	out := &limitedBuffer{limit: 4 * 1024 * 1024}
	cmd.Stdout, cmd.Stderr = out, out
	err := cmd.Run()
	if err != nil {
		return out.String(), fmt.Errorf("%s: %w\n%s", name, err, strings.TrimSpace(out.String()))
	}
	return out.String(), nil
}

type Process struct {
	PID, PPID                             int
	User, CPU                             string
	Memory                                int64
	Elapsed, Executable, Command, Started string
}

func processes(ctx context.Context, runner Runner) ([]Process, error) {
	output, err := runner(ctx, "ps", "-axo", "pid=,ppid=,user=,%cpu=,rss=,etime=,comm=")
	if err != nil {
		return nil, err
	}
	result := []Process{}
	for _, line := range strings.Split(output, "\n") {
		f := strings.Fields(line)
		if len(f) < 7 {
			continue
		}
		pid, e := strconv.Atoi(f[0])
		if e != nil {
			continue
		}
		ppid, _ := strconv.Atoi(f[1])
		rss, _ := strconv.ParseInt(f[4], 10, 64)
		result = append(result, Process{PID: pid, PPID: ppid, User: f[2], CPU: f[3], Memory: rss * 1024, Elapsed: f[5], Executable: strings.Join(f[6:], " ")})
	}
	sort.Slice(result, func(i, j int) bool {
		a, _ := strconv.ParseFloat(result[i].CPU, 64)
		b, _ := strconv.ParseFloat(result[j].CPU, 64)
		if a == b {
			return result[i].PID < result[j].PID
		}
		return a > b
	})
	return result, nil
}
func processIdentity(ctx context.Context, runner Runner, pid int) (string, error) {
	s, e := runner(ctx, "ps", "-p", strconv.Itoa(pid), "-o", "lstart=,comm=")
	if e != nil {
		return "", e
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("process has exited")
	}
	return s, nil
}
func signalProcess(ctx context.Context, runner Runner, p Process, identity string, sig syscall.Signal) error {
	if p.PID <= 1 || p.PID == os.Getpid() || p.PID == os.Getppid() {
		return errors.New("refusing to signal the manager, its parent, or system init")
	}
	current, err := processIdentity(ctx, runner, p.PID)
	if err != nil {
		return err
	}
	if current != identity {
		return errors.New("process identity changed; refresh and select again")
	}
	return syscall.Kill(p.PID, sig)
}
func inspectProcess(ctx context.Context, runner Runner, p Process) string {
	var b strings.Builder
	fmt.Fprintf(&b, "PID %d   Parent %d   User %s\nExecutable: %s\nCPU %s%%   Memory %s   Age %s\n", p.PID, p.PPID, p.User, p.Executable, p.CPU, bytesLabel(p.Memory), p.Elapsed)
	for _, task := range []struct {
		label string
		args  []string
	}{{"Command", []string{"ps", "-ww", "-p", strconv.Itoa(p.PID), "-o", "command="}}, {"Working directory and executable mappings", []string{"lsof", "-a", "-p", strconv.Itoa(p.PID), "-d", "cwd,txt", "-Fn"}}, {"Network connections", []string{"lsof", "-nP", "-a", "-p", strconv.Itoa(p.PID), "-i"}}} {
		out, err := runner(ctx, task.args[0], task.args[1:]...)
		fmt.Fprintf(&b, "\n%s\n%s", task.label, out)
		if err != nil {
			fmt.Fprintf(&b, "\nUnavailable: %v\n", err)
		}
	}
	return b.String()
}
func agentKind(p Process) string {
	base := strings.ToLower(filepath.Base(p.Executable))
	switch {
	case base == "claude" || base == "claude-code":
		return "Claude"
	case base == "codex" || strings.HasPrefix(base, "codex-"):
		return "Codex"
	case base == "go" || strings.Contains(p.Executable, "/go-build") || strings.Contains(p.Executable, "/go/bin/"):
		return "Go"
	}
	return ""
}

type Container struct {
	ID                   string
	Names                []string
	Image, State, Status string
	Ports                []struct {
		HostIP        string `json:"host_ip"`
		ContainerPort int    `json:"container_port"`
		HostPort      int    `json:"host_port"`
		Protocol      string `json:"protocol"`
	}
}

func (c Container) name() string {
	if len(c.Names) > 0 {
		return strings.Join(c.Names, ", ")
	}
	return shortID(c.ID)
}

type Image struct {
	ID         string
	RepoTags   []string
	Names      []string
	Size       int64
	Containers int
}

type Volume struct {
	Name, Driver, Mountpoint, CreatedAt string
	MountCount                          int
}

func (i Image) tags() []string {
	if len(i.RepoTags) > 0 {
		return i.RepoTags
	}
	return i.Names
}

func shortID(s string) string { s = strings.TrimPrefix(s, "sha256:"); return s[:min(12, len(s))] }
func podmanData(ctx context.Context, runner Runner) ([]Container, []Image, []Volume, error) {
	var cs []Container
	var is []Image
	var vs []Volume
	for _, q := range []struct {
		target any
		args   []string
	}{{&cs, []string{"ps", "--all", "--format", "json"}}, {&is, []string{"images", "--format", "json"}}, {&vs, []string{"volume", "ls", "--format", "json"}}} {
		s, err := runner(ctx, "podman", q.args...)
		if err != nil {
			return cs, is, vs, err
		}
		if err = json.Unmarshal([]byte(s), q.target); err != nil {
			return cs, is, vs, err
		}
	}
	return cs, is, vs, nil
}
func containerAction(ctx context.Context, runner Runner, id, action string) (string, error) {
	if id == "" || strings.HasPrefix(id, "-") {
		return "", errors.New("invalid container ID")
	}
	switch action {
	case "stop":
		return runner(ctx, "podman", "stop", "--time", "10", id)
	case "kill", "restart", "start":
		return runner(ctx, "podman", action, id)
	case "rmi":
		return runner(ctx, "podman", "rmi", "--force", id)
	case "volume-rm":
		return runner(ctx, "podman", "volume", "rm", "--force", id)
	}
	return "", errors.New("unsupported action")
}
func sshArgs(c Container, user string) ([]string, error) {
	if user == "" || strings.HasPrefix(user, "-") || strings.ContainsAny(user, " @\n\r") {
		return nil, errors.New("enter a valid SSH user")
	}
	for _, p := range c.Ports {
		if p.ContainerPort == 22 && p.Protocol == "tcp" && p.HostPort > 0 {
			host := p.HostIP
			if host == "" || host == "0.0.0.0" || host == "::" {
				host = "127.0.0.1"
			}
			return []string{"-p", strconv.Itoa(p.HostPort), "-l", user, "--", host}, nil
		}
	}
	return nil, errors.New("container must run an SSH server and publish TCP port 22; use e for a Podman shell")
}
func bytesLabel(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d B", n)
	}
	v := float64(n)
	units := []string{"kB", "MB", "GB", "TB", "PB"}
	for _, u := range units {
		v /= 1000
		if v < 1000 || u == "PB" {
			return fmt.Sprintf("%.1f %s", v, u)
		}
	}
	return "0 B"
}
