package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoveryFindsLogsWithoutTypingAPath(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	put(t, filepath.Join(dir, "open.log"), "x")
	put(t, filepath.Join(dir, "lib.dylib"), "x")
	put(t, filepath.Join(dir, "work/server.log"), "y")
	lsof := fmt.Sprintf("p42\nfcwd\ntDIR\nn%s/work\nftxt\ntREG\nn%s/lib.dylib\nf7\ntREG\nn%s/open.log\nf1\ntCHR\nn/dev/ttys001\n", dir, dir, dir)
	runner := func(context.Context, string, ...string) (string, error) { return lsof, nil }
	p := Process{PID: 42, Executable: "/usr/local/bin/api"}
	sources, err := sessionSources(ctx, runner, dir, p, false)
	if err != nil || len(sources) != 1 || sources[0].Path != dir+"/open.log" || !sources[0].Linked {
		t.Fatal("open log not found automatically", sources, err)
	}
	sources, _ = sessionSources(ctx, runner, dir, p, true)
	if len(sources) != 2 || sources[1].Path != dir+"/work/server.log" || sources[1].Linked {
		t.Fatal("l must also offer working-directory logs, marked unconfirmed", sources)
	}
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	put(t, filepath.Join(dir, ".claude/projects/other/s1.jsonl"), `{"cwd":"/elsewhere"}`)
	empty := func(context.Context, string, ...string) (string, error) { return "", nil }
	sources, err = sessionSources(ctx, empty, dir, Process{PID: 43, Executable: "claude"}, true)
	if err != nil || len(sources) != 1 || !strings.HasPrefix(sources[0].Label, "Recent Claude") || sources[0].Linked {
		t.Fatal("unlinked Claude session offered no transcript", sources, err)
	}
}
