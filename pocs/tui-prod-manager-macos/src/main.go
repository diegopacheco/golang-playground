package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	home, _ := os.UserHomeDir()
	root := flag.String("root", filepath.Join(home, "git", "diegopacheco"), "Project cleanup scan root")
	capture := flag.String("capture", "", "Write live terminal captures into this directory")
	status := flag.Bool("status", false, "Show manager instance status")
	stop := flag.Bool("stop", false, "Stop the manager instance")
	version := flag.Bool("version", false, "Print version")
	ask := flag.String("ask", "", "Ask AI a systems question without opening the TUI")
	provider := flag.String("provider", "", "AI provider override: claude, codex, ollama")
	model := flag.String("model", "", "AI model ID override")
	flag.Parse()
	if *ask != "" {
		cfg, err := loadConfig()
		if err == nil {
			if *provider != "" {
				cfg.Provider = *provider
			}
			if *model != "" {
				cfg.Model = *model
			}
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			var answer string
			answer, err = explain(ctx, run, cfg, *ask, "")
			if err == nil {
				fmt.Println(answer)
				return
			}
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *version {
		fmt.Println("Harbor 0.1.0")
		return
	}
	if *status || *stop {
		if err := manageInstance(*stop); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *capture != "" {
		if err := captureScreens(*root, *capture); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	lock, err := acquireInstance()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer lock.Close()
	u := newUI(*root)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(signals)
	go func() {
		select {
		case <-signals:
			u.cancel()
			u.app.Stop()
		case <-u.ctx.Done():
		}
	}()
	defer u.cancel()
	u.start()
	if err := u.app.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
