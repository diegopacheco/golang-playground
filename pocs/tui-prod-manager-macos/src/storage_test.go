package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func put(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
}
func TestScanGroupsProjectsWithoutDoubleCounting(t *testing.T) {
	root := t.TempDir()
	put(t, filepath.Join(root, "web/package.json"), "{}")
	put(t, filepath.Join(root, "web/node_modules/pkg/package.json"), "{}")
	put(t, filepath.Join(root, "web/node_modules/pkg/node_modules/file"), "cached")
	put(t, filepath.Join(root, "web/dist/app.js"), "built")
	put(t, filepath.Join(root, "jvm/pom.xml"), "<project/>")
	put(t, filepath.Join(root, "jvm/target/app.jar"), "built")
	put(t, filepath.Join(root, "source/target/keep"), "source")
	ps, warnings, err := scanProjects(context.Background(), root, nil)
	if err != nil || len(warnings) != 0 || len(ps) != 2 {
		t.Fatalf("%v %v %v", ps, warnings, err)
	}
	count := 0
	for _, p := range ps {
		count += len(p.Candidates)
		if p.Bytes <= 0 {
			t.Fatal(p)
		}
	}
	if count != 3 {
		t.Fatalf("nested dependencies counted twice: %d", count)
	}
}
func TestScanCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := scanProjects(ctx, t.TempDir(), nil); err == nil {
		t.Fatal("cancel ignored")
	}
}
func TestCleanupRefusesSymlinksAndTrackedFiles(t *testing.T) {
	root := t.TempDir()
	put(t, filepath.Join(root, "project/package.json"), "{}")
	put(t, filepath.Join(root, "project/node_modules/file"), "cache")
	candidate := filepath.Join(root, "project/node_modules")
	runner := func(_ context.Context, _ string, args ...string) (string, error) {
		if args[2] == "rev-parse" {
			return root, nil
		}
		return "project/node_modules/file\x00", nil
	}
	if err := validateCandidate(context.Background(), runner, root, candidate); err == nil || !strings.Contains(err.Error(), "tracked") {
		t.Fatal(err)
	}
	outside := t.TempDir()
	put(t, filepath.Join(outside, "file"), "precious")
	os.RemoveAll(candidate)
	if err := os.Symlink(outside, candidate); err != nil {
		t.Fatal(err)
	}
	if err := validateCandidate(context.Background(), runner, root, candidate); err == nil {
		t.Fatal("symlink accepted")
	}
	if err := validateCandidate(context.Background(), runner, root, outside); err == nil {
		t.Fatal("outside target accepted")
	}
}
func TestCleanupDeletesOnlyReviewedArtifacts(t *testing.T) {
	root := t.TempDir()
	put(t, filepath.Join(root, "web/package.json"), "{}")
	put(t, filepath.Join(root, "web/node_modules/file"), "cache")
	put(t, filepath.Join(root, "web/src/main.js"), "keep")
	runner := func(_ context.Context, _ string, args ...string) (string, error) {
		if args[2] == "rev-parse" {
			return root, nil
		}
		return "", nil
	}
	ps, _, err := scanProjects(context.Background(), root, nil)
	if err != nil {
		t.Fatal(err)
	}
	removed, skipped, err := reclaimProject(context.Background(), runner, root, ps[0])
	if err != nil || len(skipped) != 0 || len(removed) != 1 || removed[0].Bytes <= 0 {
		t.Fatal(removed, skipped, err)
	}
	if _, err = os.Stat(filepath.Join(root, "web/node_modules")); !os.IsNotExist(err) {
		t.Fatal("artifact remains")
	}
	if _, err = os.Stat(filepath.Join(root, "web/src/main.js")); err != nil {
		t.Fatal("source removed")
	}
}
func TestCleanupKeepsTrackedFoldersWithoutBlockingSafeOnes(t *testing.T) {
	root := t.TempDir()
	put(t, filepath.Join(root, "web/package.json"), "{}")
	put(t, filepath.Join(root, "web/node_modules/file"), "cache")
	put(t, filepath.Join(root, "web/dist/app.js"), "tracked")
	runner := func(_ context.Context, _ string, args ...string) (string, error) {
		if args[2] == "rev-parse" {
			return root, nil
		}
		if strings.HasSuffix(args[len(args)-1], "dist") {
			return "tracked", nil
		}
		return "", nil
	}
	ps, _, _ := scanProjects(context.Background(), root, nil)
	ps[0].Candidates = append(ps[0].Candidates, Candidate{filepath.Join(root, "web/.next"), 99})
	removed, skipped, err := reclaimProject(context.Background(), runner, root, ps[0])
	if err != nil || len(skipped) != 1 || !strings.Contains(skipped[0], "tracked") {
		t.Fatal("tracked folder not reported as kept", skipped, err)
	}
	if _, err := os.Stat(filepath.Join(root, "web/dist/app.js")); err != nil {
		t.Fatal("committed build output deleted")
	}
	if _, err := os.Stat(filepath.Join(root, "web/node_modules")); !os.IsNotExist(err) {
		t.Fatal("one tracked folder blocked reclaiming the safe folder")
	}
	if len(removed) != 2 || removed[1].Bytes != 0 {
		t.Fatal("already-deleted folder must leave the list without claiming freed space", removed)
	}
}

func TestCleanupUsesLiteralGitPaths(t *testing.T) {
	root := t.TempDir()
	put(t, filepath.Join(root, "web[1]/package.json"), "{}")
	put(t, filepath.Join(root, "web[1]/node_modules/owned.txt"), "source")
	ctx := context.Background()
	if _, err := run(ctx, "git", "init", root); err != nil {
		t.Fatal(err)
	}
	if _, err := run(ctx, "git", "-C", root, "add", "--", "."); err != nil {
		t.Fatal(err)
	}
	projects, _, err := scanProjects(ctx, root, nil)
	if err != nil || len(projects) != 1 {
		t.Fatal(projects, err)
	}
	_, skipped, err := reclaimProject(ctx, run, root, projects[0])
	if err != nil || len(skipped) != 1 || !strings.Contains(skipped[0], "tracked") {
		t.Fatalf("tracked path with pattern characters was not protected: %v %v", skipped, err)
	}
	if _, err = os.Stat(filepath.Join(root, "web[1]/node_modules/owned.txt")); err != nil {
		t.Fatal("tracked file deleted")
	}
}
