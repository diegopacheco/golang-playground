package shots

import (
	"errors"
	"fmt"
	"image/color"
	"reflect"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/diegopacheco/golang-playground/pocs/lipgloss-huh-fun/internal/card"
	"github.com/diegopacheco/golang-playground/pocs/lipgloss-huh-fun/internal/config"
	"github.com/diegopacheco/golang-playground/pocs/lipgloss-huh-fun/internal/wizard"
)

const cmdTimeout = 50 * time.Millisecond

var ErrNotCompleted = errors.New("scripted wizard did not complete")

type Shot struct {
	Name    string
	Title   string
	Content string
}

type driver struct {
	form *huh.Form
}

func Capture() ([]Shot, config.Config, error) {
	var answers wizard.Answers
	d := &driver{form: wizard.NewForm(&answers)}
	d.run(d.form.Init())
	d.send(tea.WindowSizeMsg{Width: 100, Height: 30})
	d.send(tea.BackgroundColorMsg{Color: color.Black})

	var shots []Shot
	snap := func(name, title string) {
		shots = append(shots, Shot{Name: name, Title: title, Content: d.form.View()})
	}

	snap("01-name", "Step 1: project name")
	d.typeText("Billing API")
	d.key(tea.KeyEnter)
	snap("02-name-invalid", "Step 1: validation rejects an invalid name")
	for range len("Billing API") {
		d.key(tea.KeyBackspace)
	}
	d.typeText("billing-api")
	snap("03-name-valid", "Step 1: a valid kebab-case name")
	d.key(tea.KeyEnter)
	d.key(tea.KeyDown)
	snap("04-stack", "Step 2: stack select")
	d.key(tea.KeyEnter)
	d.key(tea.KeySpace)
	d.key(tea.KeyDown)
	d.key(tea.KeySpace)
	d.key(tea.KeyDown)
	d.key(tea.KeyDown)
	d.key(tea.KeySpace)
	snap("05-features", "Step 3: features multi-select")
	d.key(tea.KeyEnter)
	d.key(tea.KeyLeft)
	snap("06-confirm", "Step 4: confirm")
	d.key(tea.KeyEnter)

	if d.form.State != huh.StateCompleted || !answers.Confirmed {
		return nil, config.Config{}, fmt.Errorf("%w: state %v", ErrNotCompleted, d.form.State)
	}
	cfg := answers.Config()
	shots = append(shots, Shot{Name: "07-summary", Title: "Summary card", Content: card.Render(cfg)})

	_, err := config.Parse("Billing API", "cobol", "docker,k8s")
	if err == nil {
		return nil, config.Config{}, errors.New("invalid flags were accepted")
	}
	shots = append(shots, Shot{Name: "08-flags-invalid", Title: "Non-interactive mode with invalid flags", Content: card.RenderError(err)})
	return shots, cfg, nil
}

func (d *driver) typeText(s string) {
	for _, r := range s {
		d.send(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
}

func (d *driver) key(code rune) {
	msg := tea.KeyPressMsg{Code: code}
	if code == tea.KeySpace {
		msg.Text = " "
	}
	d.send(msg)
}

func (d *driver) send(msg tea.Msg) {
	m, cmd := d.form.Update(msg)
	d.form = m.(*huh.Form)
	d.run(cmd)
}

func (d *driver) run(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	msg, ok := resolve(cmd)
	if !ok || msg == nil {
		return
	}
	if cmds, isList := asCmds(msg); isList {
		for _, c := range cmds {
			d.run(c)
		}
		return
	}
	d.send(msg)
}

func resolve(cmd tea.Cmd) (tea.Msg, bool) {
	ch := make(chan tea.Msg, 1)
	go func() { ch <- cmd() }()
	select {
	case msg := <-ch:
		return msg, true
	case <-time.After(cmdTimeout):
		return nil, false
	}
}

func asCmds(msg tea.Msg) ([]tea.Cmd, bool) {
	v := reflect.ValueOf(msg)
	cmdType := reflect.TypeFor[tea.Cmd]()
	if v.Kind() != reflect.Slice || v.Type().Elem() != cmdType {
		return nil, false
	}
	cmds := make([]tea.Cmd, v.Len())
	for i := range cmds {
		cmds[i] = v.Index(i).Interface().(tea.Cmd)
	}
	return cmds, true
}
