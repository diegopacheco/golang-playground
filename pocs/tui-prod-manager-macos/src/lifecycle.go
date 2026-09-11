package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

type Instance struct {
	PID      int
	Identity string
}

func lockPath() string { return filepath.Join(filepath.Dir(configPath()), "instance.lock") }
func acquireInstance() (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(lockPath()), 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(lockPath(), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, errors.New("Harbor is already running; use harbor --status")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	identity, err := processIdentity(ctx, run, os.Getpid())
	if err != nil {
		f.Close()
		return nil, err
	}
	data, err := json.Marshal(Instance{os.Getpid(), identity})
	if err != nil {
		f.Close()
		return nil, err
	}
	if err = f.Truncate(0); err != nil {
		f.Close()
		return nil, err
	}
	if _, err = f.WriteAt(data, 0); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}
func manageInstance(stop bool) error {
	f, err := os.OpenFile(lockPath(), os.O_RDWR, 0600)
	if os.IsNotExist(err) {
		fmt.Println("Harbor DOWN")
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	if syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) == nil {
		fmt.Println("Harbor DOWN")
		return nil
	}
	var current Instance
	if err = json.NewDecoder(f).Decode(&current); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	identity, err := processIdentity(ctx, run, current.PID)
	if err != nil {
		return err
	}
	if identity != current.Identity {
		return errors.New("instance PID identity does not match")
	}
	if !stop {
		fmt.Printf("Harbor UP pid=%d\n", current.PID)
		return nil
	}
	if err = syscall.Kill(current.PID, syscall.SIGTERM); err != nil {
		return err
	}
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return errors.New("Harbor did not stop within 10 seconds")
		case <-ticker.C:
			if syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) == nil {
				fmt.Println("Harbor stopped")
				return nil
			}
		}
	}
}
