# Design Doc - Go Playground

## Goal
A local admin UI to write and run Go code with line numbers, syntax highlighting and auto-complete, plus rich markdown docs and runnable code samples for the core Go concurrency features: goroutines, channels, buffered channels, select, sync, context, worker pools and pipelines. It ships as a web app and as a macOS Electron app.

## Architecture
- A single Go binary (`cmd/server`) serves the JSON API and the static UI. The UI files, the markdown docs and the code samples are embedded with `embed`, so one binary holds everything.
- `internal/runner` writes the submitted code into a fresh temp dir and runs `go run main.go` with a hard timeout and an output cap. The temp dir is removed after every run.
- `internal/complete` builds an index at startup: Go keywords, builtins, and the exported identifiers of a fixed list of standard packages parsed from `GOROOT/src` with `go/parser`. Queries are prefix matches.
- `internal/markdown` renders the docs markdown to HTML on the server (headings, paragraphs, lists, tables, code blocks, inline code, bold, italics, links, callouts).
- `internal/content` lists the embedded docs and samples.
- `internal/api` holds the HTTP handlers and maps errors to status codes in one place.
- `internal/logx` is the only logger, adding correlation id, path, user id and service name.
- `web/` is a vanilla JS UI: a textarea with a highlighted `<pre>` overlay and a line number gutter, a completion popup, and tabs for Playground, Docs and Samples.
- `electron/` wraps the UI. On start it runs `scripts/start-all.sh`, shows a boot screen with every service status, then loads the UI. On quit it runs `scripts/stop-all.sh`.

```
Electron window / Browser
        |
        v
  Go server :8086  --- embedded web UI, docs, samples
        |
   +----+-----------+-------------+
   |                |             |
 runner          complete      markdown
 (go run in      (stdlib       (docs to HTML)
  temp dir)       index)
```

## Stack
- Go 1.26 standard library only: `net/http`, `go/parser`, `go/format`, `embed`, `log/slog`, `os/exec`.
- Vanilla HTML, CSS and JS for the UI, no frameworks and no editor library.
- Electron for the macOS app, packaged with `@electron/packager`.
- Bash scripts for setup, start, stop, status, test, ui, install and uninstall.

## Public Endpoints
| Method | Path | Body | Response |
|---|---|---|---|
| GET | `/api/health` | - | `{"status":"ok","services":[{"name","status"}]}` |
| POST | `/api/run` | `{"code"}` | `{"stdout","stderr","exitCode","durationMs","timedOut"}` |
| POST | `/api/format` | `{"code"}` | `{"code"}` or 400 with `{"error"}` |
| GET | `/api/complete?prefix=fmt.Pr` | - | `[{"label","kind","detail"}]` |
| GET | `/api/docs` | - | `[{"id","title","summary"}]` |
| GET | `/api/docs/{id}` | - | `{"id","title","html"}` |
| GET | `/api/samples` | - | `[{"id","title","code"}]` |
| GET | `/` | - | UI |

## Important Design Decisions
- Code runs with the local `go` toolchain in a temp dir with a 10s timeout and a 64KB output cap. It is a local tool and has no sandbox, so the server binds to `127.0.0.1` only.
- No editor library. A textarea under a highlighted overlay keeps native copy, paste, cut, undo and selection, and keeps the UI dependency free.
- Completions come from parsing the real standard library sources at startup, so they always match the installed Go version.
- Markdown is rendered on the server so it has Go tests. Code blocks in docs get a "Run in Playground" button.
- Every sample is a complete `main` package and a test runs each one, so a broken sample fails the build.

## Pros
- One binary, no runtime dependencies besides Go itself.
- Completions and execution always match the installed Go version.
- Same UI for web, mobile viewport and Electron.

## Cons
- No sandbox: code runs with the user's permissions. Fine for a local tool, not for a public server.
- Completion is prefix based, not type aware like gopls.
- The overlay editor is simpler than a full editor: no multi-cursor and no code folding.
