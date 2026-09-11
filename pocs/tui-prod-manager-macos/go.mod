module github.com/diegopacheco/tui-prod-manager-macos

go 1.26

replace github.com/diegopacheco/agent-sdk-go => ../../personal-private-software/agents-sdks/go

require (
	github.com/diegopacheco/agent-sdk-go v0.0.0-00010101000000-000000000000
	github.com/gdamore/tcell/v2 v2.7.1
	github.com/rivo/tview v0.0.0-20241227133733-17b7edb88c57
)

require (
	github.com/gdamore/encoding v1.0.0
	github.com/lucasb-eyer/go-colorful v1.2.0
	github.com/mattn/go-runewidth v0.0.15
	github.com/rivo/uniseg v0.4.7
	golang.org/x/sys v0.17.0
	golang.org/x/term v0.17.0
	golang.org/x/text v0.14.0
)
