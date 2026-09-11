package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

type Candidate struct {
	Path  string
	Bytes int64
}
type Project struct {
	Path       string
	Candidates []Candidate
	Bytes      int64
}
type DiskEntry struct {
	Path  string
	Bytes int64
}
type Storage struct {
	Apps, Folders []DiskEntry
	Total, Free   uint64
	Warnings      []string
}

func hasFile(dir string, names ...string) bool {
	for _, n := range names {
		if s, e := os.Stat(filepath.Join(dir, n)); e == nil && !s.IsDir() {
			return true
		}
	}
	return false
}
func reclaimable(path string) bool {
	parent := filepath.Dir(path)
	switch filepath.Base(path) {
	case "node_modules":
		return hasFile(parent, "package.json")
	case "target":
		return hasFile(parent, "pom.xml", "build.sbt", "Cargo.toml", "build.sc")
	case "build":
		return hasFile(parent, "build.gradle", "build.gradle.kts", "CMakeLists.txt")
	case "dist":
		return hasFile(parent, "package.json")
	case ".next":
		return hasFile(parent, "package.json")
	}
	return false
}
func diskSize(ctx context.Context, path string) (int64, error) {
	var size int64
	var first error
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, e error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if e != nil {
			if first == nil {
				first = e
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		s, e := d.Info()
		if e != nil {
			if first == nil {
				first = e
			}
			return nil
		}
		if st, ok := s.Sys().(*syscall.Stat_t); ok {
			size += st.Blocks * 512
		} else {
			size += s.Size()
		}
		return nil
	})
	if err != nil {
		return size, err
	}
	return size, first
}
func scanProjects(ctx context.Context, root string, progress func(string)) ([]Project, []string, error) {
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, nil, err
	}
	found := map[string]*Project{}
	warnings := []string{}
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if e != nil {
			if len(warnings) < 20 {
				warnings = append(warnings, e.Error())
			}
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if d.Name() == ".git" {
			return filepath.SkipDir
		}
		if !reclaimable(path) {
			return nil
		}
		if progress != nil {
			progress(path)
		}
		size, e := diskSize(ctx, path)
		if e != nil {
			if len(warnings) < 20 {
				warnings = append(warnings, e.Error())
			}
			return filepath.SkipDir
		}
		project := filepath.Dir(path)
		if found[project] == nil {
			found[project] = &Project{Path: project}
		}
		found[project].Candidates = append(found[project].Candidates, Candidate{path, size})
		found[project].Bytes += size
		return filepath.SkipDir
	})
	result := []Project{}
	for _, p := range found {
		result = append(result, *p)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Bytes > result[j].Bytes })
	return result, warnings, err
}
func validateCandidate(ctx context.Context, runner Runner, root, path string) error {
	originalRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	root, err = filepath.EvalSymlinks(originalRoot)
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(originalRoot, abs)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		rel, err = filepath.Rel(root, abs)
		if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return errors.New("cleanup target is outside the scan root")
		}
	}
	expected := filepath.Join(root, rel)
	actual, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return err
	}
	if actual != expected {
		return errors.New("cleanup refuses symbolic links")
	}
	abs = actual
	if !reclaimable(abs) {
		return errors.New("target is no longer a recognized build or dependency directory")
	}
	info, err := os.Stat(abs)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("cleanup target is not a directory")
	}
	repo, err := runner(ctx, "git", "-C", filepath.Dir(abs), "rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Errorf("cannot verify Git ownership: %w", err)
	}
	relative, err := filepath.Rel(strings.TrimSpace(repo), abs)
	if err != nil {
		return err
	}
	tracked, err := runner(ctx, "git", "-C", strings.TrimSpace(repo), "ls-files", "-z", "--", ":(literal)"+relative)
	if err != nil {
		return err
	}
	if tracked != "" {
		return errors.New("cleanup refuses directories containing Git-tracked files")
	}
	return nil
}
func reclaimProject(ctx context.Context, runner Runner, root string, p Project) ([]Candidate, []string, error) {
	var removed []Candidate
	var skipped []string
	for _, c := range p.Candidates {
		if ctx.Err() != nil {
			return removed, skipped, ctx.Err()
		}
		if _, err := os.Lstat(c.Path); os.IsNotExist(err) {
			removed = append(removed, Candidate{c.Path, 0})
			continue
		}
		if err := validateCandidate(ctx, runner, root, c.Path); err != nil {
			skipped = append(skipped, c.Path+": "+err.Error())
			continue
		}
		quarantine, err := os.MkdirTemp(filepath.Dir(c.Path), ".harbor-reclaim-")
		if err != nil {
			return removed, skipped, err
		}
		moved := filepath.Join(quarantine, filepath.Base(c.Path))
		if err = os.Rename(c.Path, moved); err != nil {
			os.Remove(quarantine)
			return removed, skipped, err
		}
		if err = os.RemoveAll(quarantine); err != nil {
			return removed, skipped, fmt.Errorf("partial cleanup remains at %s: %w", quarantine, err)
		}
		removed = append(removed, c)
	}
	return removed, skipped, nil
}
func scanStorage(ctx context.Context, home string, progress func(string)) Storage {
	s := Storage{}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(home, &stat); err == nil {
		s.Total = stat.Blocks * uint64(stat.Bsize)
		s.Free = stat.Bavail * uint64(stat.Bsize)
	} else {
		s.Warnings = append(s.Warnings, err.Error())
	}
	collect := func(root string, apps bool) {
		entries, err := os.ReadDir(root)
		if err != nil {
			if !os.IsNotExist(err) {
				s.Warnings = append(s.Warnings, err.Error())
			}
			return
		}
		for _, e := range entries {
			if ctx.Err() != nil {
				return
			}
			if !e.IsDir() || e.Type()&os.ModeSymlink != 0 {
				continue
			}
			if apps && !strings.HasSuffix(e.Name(), ".app") {
				continue
			}
			path := filepath.Join(root, e.Name())
			if progress != nil {
				progress(path)
			}
			size, err := diskSize(ctx, path)
			if err != nil && len(s.Warnings) < 20 {
				s.Warnings = append(s.Warnings, err.Error())
			}
			entry := DiskEntry{path, size}
			if apps {
				s.Apps = append(s.Apps, entry)
			} else {
				s.Folders = append(s.Folders, entry)
			}
		}
	}
	collect("/Applications", true)
	collect("/System/Applications", true)
	collect(filepath.Join(home, "Applications"), true)
	collect(home, false)
	sort.Slice(s.Apps, func(i, j int) bool { return s.Apps[i].Bytes > s.Apps[j].Bytes })
	sort.Slice(s.Folders, func(i, j int) bool { return s.Folders[i].Bytes > s.Folders[j].Bytes })
	s.Apps = s.Apps[:min(10, len(s.Apps))]
	s.Folders = s.Folders[:min(10, len(s.Folders))]
	if ctx.Err() != nil {
		s.Warnings = append(s.Warnings, ctx.Err().Error())
	}
	return s
}
