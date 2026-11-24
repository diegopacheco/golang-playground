package tail

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"
)

type Line struct {
	File string
	Text string
}

type Options struct {
	Interval  time.Duration
	FromStart bool
}

type follower struct {
	path    string
	file    *os.File
	info    os.FileInfo
	offset  int64
	partial []byte
	buf     []byte
	out     chan<- Line
}

func Follow(ctx context.Context, path string, opts Options, out chan<- Line) error {
	f := &follower{path: path, out: out, buf: make([]byte, 32*1024)}
	defer f.close()
	if err := f.open(opts.FromStart); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()
	for {
		if err := f.poll(ctx); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (f *follower) open(fromStart bool) error {
	file, err := os.Open(f.path)
	if err != nil {
		return fmt.Errorf("open %s: %w", f.path, err)
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("stat %s: %w", f.path, err)
	}
	f.file, f.info, f.offset, f.partial = file, info, 0, nil
	if !fromStart {
		f.offset, err = file.Seek(0, io.SeekEnd)
		if err != nil {
			f.close()
			return fmt.Errorf("seek %s: %w", f.path, err)
		}
	}
	return nil
}

func (f *follower) close() {
	if f.file != nil {
		f.file.Close()
		f.file = nil
	}
}

func (f *follower) poll(ctx context.Context) error {
	if f.file == nil {
		if err := f.open(true); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
	}
	if err := f.drain(ctx); err != nil {
		return err
	}
	current, err := os.Stat(f.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat %s: %w", f.path, err)
	}
	if !os.SameFile(f.info, current) {
		f.close()
		if err := f.open(true); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		return f.drain(ctx)
	}
	if current.Size() < f.offset {
		if _, err := f.file.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("seek %s: %w", f.path, err)
		}
		f.offset, f.partial = 0, nil
		return f.drain(ctx)
	}
	return nil
}

func (f *follower) drain(ctx context.Context) error {
	for {
		n, err := f.file.Read(f.buf)
		if n > 0 {
			f.offset += int64(n)
			if sendErr := f.emit(ctx, f.buf[:n]); sendErr != nil {
				return sendErr
			}
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read %s: %w", f.path, err)
		}
	}
}

func (f *follower) emit(ctx context.Context, chunk []byte) error {
	f.partial = append(f.partial, chunk...)
	for {
		i := bytes.IndexByte(f.partial, '\n')
		if i < 0 {
			return nil
		}
		text := string(bytes.TrimSuffix(f.partial[:i], []byte("\r")))
		f.partial = f.partial[i+1:]
		select {
		case f.out <- Line{File: f.path, Text: text}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
