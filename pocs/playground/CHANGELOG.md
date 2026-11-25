# Changelog

All notable changes to Go Playground.

## 1.0.0 - 2026-11-25

### Added
- A Playground tab with a Go editor that has line numbers, syntax highlighting, auto-indent, bracket pairs and auto-complete from the installed standard library.
- Run (⌘ Enter) builds and runs the code, then shows stdout, stderr, the real exit code and the duration. Runs stop after 10 seconds and output stops at 64KB.
- Format (⇧ ⌘ F) rewrites the code with gofmt.
- Compile errors turn their line red in the gutter.
- A Docs tab with 11 articles on goroutines, channels, buffered channels, select, sync, atomics, context, worker pools, pipelines, defer/panic/recover and generics. Each code block has Copy and "Run in Playground" buttons.
- A Samples tab with 13 runnable programs that open or run in one click.
- Search (⌘ K) over tabs, docs, samples and actions, and a shortcuts modal (⌘ /) with a filter box.
- Tabs on ⌘ 1-3, zoom on ⌘ +, ⌘ - and ⌘ 0, full screen on ⇧ ⌘ Enter, and hover tooltips that show each shortcut.
- Every panel can expand to almost full screen or hide to a side tab, and its hidden state is restored on the next launch.
- A macOS app with a boot screen. It starts the server on launch, stops it on quit, allows only one instance, and remembers its window position. Double-clicking the title bar maximizes it, and ⌘ P saves a screenshot to the Desktop.
- Scripts for setup, start, stop, status, tests, opening the UI, and installing or uninstalling the macOS app.

### Verified
- Go: gofmt and go vet are clean. All 5 tested packages pass with 0 skips, and every one of the 13 samples and every runnable doc block runs with exit code 0.
- Web: 24 of 24 tests pass. Electron: 6 of 6 tests pass.
- In headless Chromium, with no console errors: Run printed the goroutine output with exit 0, typing `sync.Wa` suggested `WaitGroup` and Tab inserted it, ⌘ K and ⌘ / filtered results and closed on Esc, hiding a panel docked it to the right edge, and a compile error marked line 3.
- On a 440px-wide phone screen there is no horizontal scroll and no button is under 44px.
- The installed macOS app started the server in 3.5s, opening it a second time kept 1 process, and quitting stopped the server.
