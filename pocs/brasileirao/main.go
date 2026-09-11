package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gdamore/tcell/v2"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/rivo/tview"
)

type TeamData struct {
	Position  string
	Team      string
	Points    string
	Games     string
	Victories string
	Draws     string
	Losses    string
	GP        string
	GC        string
	SG        string
	Percent   string
	Last5     string
}

type CachedData struct {
	Timestamp time.Time
	Teams     []TeamData
}

type ProgressUpdate struct {
	Phase   string
	Message string
}

func loadCache() ([]TeamData, bool) {
	data, err := os.ReadFile("cache.txt")
	if err != nil {
		return nil, false
	}

	var cached CachedData
	err = json.Unmarshal(data, &cached)
	if err != nil {
		return nil, false
	}

	if time.Since(cached.Timestamp) < 2*time.Hour {
		log.Printf("Using cached data from %v\n", cached.Timestamp)
		return cached.Teams, true
	}

	log.Println("Cache expired")
	return nil, false
}

func saveCache(teams []TeamData) {
	cached := CachedData{
		Timestamp: time.Now(),
		Teams:     teams,
	}

	data, err := json.Marshal(cached)
	if err != nil {
		log.Printf("Failed to marshal cache: %v\n", err)
		return
	}

	err = os.WriteFile("cache.txt", data, 0644)
	if err != nil {
		log.Printf("Failed to write cache: %v\n", err)
		return
	}

	log.Println("Cache saved")
}

func fetchBrasileiraoTable(progressChan chan<- ProgressUpdate) ([]TeamData, error) {
	if teams, ok := loadCache(); ok {
		if progressChan != nil {
			progressChan <- ProgressUpdate{"cache", "Loaded from cache"}
			close(progressChan)
		}
		return teams, nil
	}
	url := "https://ge.globo.com/futebol/brasileirao-serie-a/"

	if progressChan != nil {
		progressChan <- ProgressUpdate{"browser", "Launching browser..."}
	}
	log.Println("Launching browser...")
	l := launcher.New().Headless(true).Leakless(false)
	controlURL, err := l.Launch()
	if err != nil {
		log.Printf("Failed to launch browser: %v\n", err)
		if progressChan != nil {
			close(progressChan)
		}
		return getFallbackData(), nil
	}

	if progressChan != nil {
		progressChan <- ProgressUpdate{"connect", "Connecting to browser..."}
	}
	log.Println("Connecting to browser...")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	browser := rod.New().ControlURL(controlURL).Context(ctx)
	err = browser.Connect()
	if err != nil {
		log.Printf("Failed to connect to browser: %v\n", err)
		l.Cleanup()
		if progressChan != nil {
			close(progressChan)
		}
		return getFallbackData(), nil
	}

	if progressChan != nil {
		progressChan <- ProgressUpdate{"page", "Creating page..."}
	}
	log.Println("Creating page...")
	page, err := browser.Page(proto.TargetCreateTarget{URL: url})
	if err != nil {
		log.Printf("Failed to create page: %v\n", err)
		browser.Close()
		l.Cleanup()
		if progressChan != nil {
			close(progressChan)
		}
		return getFallbackData(), nil
	}

	if progressChan != nil {
		progressChan <- ProgressUpdate{"navigate", "Navigating to GE Globo..."}
	}
	log.Println("Navigating to page...")
	err = page.Navigate(url)
	if err != nil {
		log.Printf("Failed to navigate: %v\n", err)
		page.Close()
		browser.Close()
		l.Cleanup()
		if progressChan != nil {
			close(progressChan)
		}
		return getFallbackData(), nil
	}

	if progressChan != nil {
		progressChan <- ProgressUpdate{"load", "Waiting for page to load..."}
	}
	log.Println("Waiting for page to load...")
	err = page.WaitLoad()
	if err != nil {
		log.Printf("Failed to wait for page load: %v\n", err)
		page.Close()
		browser.Close()
		l.Cleanup()
		if progressChan != nil {
			close(progressChan)
		}
		return getFallbackData(), nil
	}

	if progressChan != nil {
		progressChan <- ProgressUpdate{"render", "Waiting for JavaScript to render..."}
	}
	log.Println("Waiting for JavaScript to render...")
	time.Sleep(5 * time.Second)

	if progressChan != nil {
		progressChan <- ProgressUpdate{"extract", "Extracting table data..."}
	}
	log.Println("Extracting HTML...")
	html, err := page.HTML()
	if err != nil {
		log.Printf("Failed to get HTML: %v\n", err)
		page.Close()
		browser.Close()
		l.Cleanup()
		if progressChan != nil {
			close(progressChan)
		}
		return getFallbackData(), nil
	}

	page.Close()
	browser.Close()
	l.Cleanup()
	log.Println("Browser closed")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Printf("Failed to parse HTML: %v\n", err)
		return getFallbackData(), nil
	}

	var teams []TeamData

	tabelaSection := doc.Find(".tabela.tabela__pontos-corridos")
	if tabelaSection.Length() == 0 {
		return getFallbackData(), nil
	}

	tabelaEquipes := tabelaSection.Find("table.tabela__equipes tbody tr")
	tabelaPontos := tabelaSection.Find("table.tabela__pontos tbody tr")

	tabelaEquipes.Each(func(idx int, row *goquery.Selection) {
		team := TeamData{}
		team.Position = strings.TrimSpace(row.Find(".classificacao__equipes--posicao").Text())
		team.Team = strings.TrimSpace(row.Find(".classificacao__equipes--nome").Text())

		if team.Team != "" && team.Position != "" {
			teams = append(teams, team)
		}
	})

	tabelaPontos.Each(func(idx int, row *goquery.Selection) {
		if idx < len(teams) {
			cells := row.Find("td.classificacao__pontos")

			if cells.Length() >= 10 {
				teams[idx].Points = strings.TrimSpace(cells.Eq(0).Text())
				teams[idx].Games = strings.TrimSpace(cells.Eq(1).Text())
				teams[idx].Victories = strings.TrimSpace(cells.Eq(2).Text())
				teams[idx].Draws = strings.TrimSpace(cells.Eq(3).Text())
				teams[idx].Losses = strings.TrimSpace(cells.Eq(4).Text())
				teams[idx].GP = strings.TrimSpace(cells.Eq(5).Text())
				teams[idx].GC = strings.TrimSpace(cells.Eq(6).Text())
				teams[idx].SG = strings.TrimSpace(cells.Eq(7).Text())
				teams[idx].Percent = strings.TrimSpace(cells.Eq(8).Text())

				lastGames := ""
				cells.Eq(9).Find(".classificacao__ultimos_jogos").Each(func(k int, span *goquery.Selection) {
					class, _ := span.Attr("class")
					if strings.Contains(class, "--v") {
						lastGames += "V"
					} else if strings.Contains(class, "--e") {
						lastGames += "E"
					} else if strings.Contains(class, "--d") {
						lastGames += "D"
					}
				})
				teams[idx].Last5 = lastGames
			}
		}
	})

	if len(teams) == 0 || teams[0].Points == "" {
		log.Println("No data found, using fallback data")
		if progressChan != nil {
			close(progressChan)
		}
		return getFallbackData(), nil
	}

	log.Printf("Successfully fetched %d teams\n", len(teams))
	saveCache(teams)

	if progressChan != nil {
		progressChan <- ProgressUpdate{"done", "Data loaded successfully!"}
		close(progressChan)
	}

	return teams, nil
}

func getFallbackData() []TeamData {
	return []TeamData{
		{"1", "Botafogo", "76", "38", "23", "7", "8", "59", "29", "30", "66.7", "VVEVV"},
		{"2", "Palmeiras", "73", "38", "22", "7", "9", "61", "26", "35", "64.0", "EVVVV"},
		{"3", "Flamengo", "70", "38", "21", "7", "10", "60", "42", "18", "61.4", "VEVVD"},
		{"4", "Fortaleza", "68", "38", "20", "8", "10", "53", "39", "14", "59.6", "VVEVV"},
		{"5", "Internacional", "69", "38", "20", "9", "9", "53", "36", "17", "60.5", "DVVVV"},
		{"6", "São Paulo", "63", "38", "19", "6", "13", "53", "43", "10", "55.3", "VEVVD"},
		{"7", "Corinthians", "56", "38", "17", "5", "16", "56", "52", "4", "49.1", "VVVVE"},
		{"8", "Bahia", "53", "38", "15", "8", "15", "49", "49", "0", "46.5", "DVEDV"},
		{"9", "Cruzeiro", "52", "38", "14", "10", "14", "43", "43", "0", "45.6", "EDVVD"},
		{"10", "Vasco", "50", "38", "14", "8", "16", "41", "56", "-15", "43.9", "DVVDV"},
		{"11", "Vitória", "46", "38", "13", "7", "18", "45", "52", "-7", "40.4", "VDDVD"},
		{"12", "Atlético-MG", "47", "38", "13", "8", "17", "47", "54", "-7", "41.2", "VDEVV"},
		{"13", "Grêmio", "45", "38", "12", "9", "17", "44", "50", "-6", "39.5", "VDEDV"},
		{"14", "Juventude", "43", "38", "12", "7", "19", "48", "59", "-11", "37.7", "DVEDD"},
		{"15", "Fluminense", "43", "38", "12", "7", "19", "33", "39", "-6", "37.7", "VEVDD"},
		{"16", "Criciúma", "42", "38", "11", "9", "18", "42", "61", "-19", "36.8", "VDEDD"},
		{"17", "Bragantino", "41", "38", "11", "8", "19", "43", "55", "-12", "35.9", "DEDDV"},
		{"18", "Athletico-PR", "42", "38", "11", "9", "18", "40", "51", "-11", "36.8", "DVDVV"},
		{"19", "Cuiabá", "30", "38", "7", "9", "22", "27", "48", "-21", "26.3", "EDDDD"},
		{"20", "Atlético-GO", "27", "38", "7", "6", "25", "29", "66", "-37", "23.7", "DDDDD"},
	}
}

func main() {
	app := tview.NewApplication()

	progressText := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	progressBox := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(progressText, 10, 0, false).
		AddItem(tview.NewBox(), 0, 1, false)

	progressChan := make(chan ProgressUpdate, 10)

	var teams []TeamData
	var fetchErr error

	go func() {
		teams, fetchErr = fetchBrasileiraoTable(progressChan)
		app.QueueUpdateDraw(func() {})
	}()

	go func() {
		phases := []string{
			"[yellow]⏳[white] Checking cache...",
		}
		currentPhase := 0
		progressText.SetText(strings.Join(phases, "\n"))

		for update := range progressChan {
			switch update.Phase {
			case "cache":
				phases = []string{
					"[green]✓[white] " + update.Message,
				}
			case "browser":
				phases = []string{
					"[green]✓[white] Checking cache...",
					"[yellow]⏳[white] " + update.Message,
				}
			case "connect":
				phases[1] = "[green]✓[white] Browser launched"
				phases = append(phases, "[yellow]⏳[white] "+update.Message)
			case "page":
				phases[2] = "[green]✓[white] Browser connected"
				phases = append(phases, "[yellow]⏳[white] "+update.Message)
			case "navigate":
				phases[3] = "[green]✓[white] Page created"
				phases = append(phases, "[yellow]⏳[white] "+update.Message)
			case "load":
				phases[4] = "[green]✓[white] Navigated to GE Globo"
				phases = append(phases, "[yellow]⏳[white] "+update.Message)
			case "render":
				phases[5] = "[green]✓[white] Page loaded"
				phases = append(phases, "[yellow]⏳[white] "+update.Message)
			case "extract":
				phases[6] = "[green]✓[white] JavaScript rendered"
				phases = append(phases, "[yellow]⏳[white] "+update.Message)
			case "done":
				phases[len(phases)-1] = "[green]✓[white] Table data extracted"
				phases = append(phases, "[green]✓[white] "+update.Message)
			}
			currentPhase++
			progressText.SetText(strings.Join(phases, "\n"))
			app.QueueUpdateDraw(func() {})
		}

		time.Sleep(500 * time.Millisecond)
		app.QueueUpdateDraw(func() {
			showMainTable(app, teams)
		})
	}()

	if fetchErr != nil {
		log.Fatalf("Error fetching table: %v", fetchErr)
	}

	if err := app.SetRoot(progressBox, true).Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}

func createTable(teams []TeamData, startIdx, endIdx int) *tview.Table {
	table := tview.NewTable().
		SetBorders(true).
		SetFixed(1, 0).
		SetSeparator('│')

	headers := []string{"Pos", "Team", "Pts", "J", "V", "E", "D", "GP", "GC", "SG", "%", "Last 5"}
	for col, header := range headers {
		table.SetCell(0, col,
			tview.NewTableCell(header).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignCenter).
				SetSelectable(false))
	}

	for idx := startIdx; idx < endIdx && idx < len(teams); idx++ {
		team := teams[idx]
		row := idx - startIdx + 1

		bgColor := tcell.ColorDefault
		textColor := tcell.ColorWhite

		if idx < 4 {
			bgColor = tcell.NewRGBColor(0, 100, 0)
			textColor = tcell.ColorWhite
		} else if idx >= 16 {
			bgColor = tcell.NewRGBColor(139, 0, 0)
			textColor = tcell.ColorWhite
		}

		last5Display := ""
		for _, char := range team.Last5 {
			switch char {
			case 'V':
				last5Display += "🟢"
			case 'D':
				last5Display += "🔴"
			case 'E':
				last5Display += "⚪"
			}
		}

		table.SetCell(row, 0, tview.NewTableCell(team.Position).SetAlign(tview.AlignCenter).SetBackgroundColor(bgColor).SetTextColor(textColor))
		table.SetCell(row, 1, tview.NewTableCell(team.Team).SetAlign(tview.AlignLeft).SetBackgroundColor(bgColor).SetTextColor(textColor))
		table.SetCell(row, 2, tview.NewTableCell(team.Points).SetAlign(tview.AlignCenter).SetBackgroundColor(bgColor).SetTextColor(textColor))
		table.SetCell(row, 3, tview.NewTableCell(team.Games).SetAlign(tview.AlignCenter).SetBackgroundColor(bgColor).SetTextColor(textColor))
		table.SetCell(row, 4, tview.NewTableCell(team.Victories).SetAlign(tview.AlignCenter).SetBackgroundColor(bgColor).SetTextColor(textColor))
		table.SetCell(row, 5, tview.NewTableCell(team.Draws).SetAlign(tview.AlignCenter).SetBackgroundColor(bgColor).SetTextColor(textColor))
		table.SetCell(row, 6, tview.NewTableCell(team.Losses).SetAlign(tview.AlignCenter).SetBackgroundColor(bgColor).SetTextColor(textColor))
		table.SetCell(row, 7, tview.NewTableCell(team.GP).SetAlign(tview.AlignCenter).SetBackgroundColor(bgColor).SetTextColor(textColor))
		table.SetCell(row, 8, tview.NewTableCell(team.GC).SetAlign(tview.AlignCenter).SetBackgroundColor(bgColor).SetTextColor(textColor))
		table.SetCell(row, 9, tview.NewTableCell(team.SG).SetAlign(tview.AlignCenter).SetBackgroundColor(bgColor).SetTextColor(textColor))
		table.SetCell(row, 10, tview.NewTableCell(team.Percent).SetAlign(tview.AlignCenter).SetBackgroundColor(bgColor).SetTextColor(textColor))
		table.SetCell(row, 11, tview.NewTableCell(last5Display).SetAlign(tview.AlignCenter).SetBackgroundColor(bgColor).SetTextColor(textColor))
	}

	table.SetSelectable(true, false).
		SetSelectedStyle(tcell.StyleDefault.
			Background(tcell.ColorYellow).
			Foreground(tcell.ColorBlack))

	return table
}

func showMainTable(app *tview.Application, teams []TeamData) {
	table1 := createTable(teams, 0, 10)
	table2 := createTable(teams, 10, 20)

	legend := tview.NewTextView().
		SetText("Campeonato Brasileiro 2025 - Serie A | G4: Libertadores | Z4: Relegation").
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorGreen)

	footer := tview.NewTextView().
		SetText("Press 'q' to quit | Arrow keys to navigate | 🟢 Victory  ⚪ Draw  🔴 Defeat").
		SetTextAlign(tview.AlignCenter)

	tablesRow := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(table1, 0, 1, true).
		AddItem(table2, 0, 1, false)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(legend, 1, 0, false).
		AddItem(tablesRow, 0, 1, true).
		AddItem(footer, 1, 0, false)

	globalPosition := 1

	updateSelection := func() {
		if globalPosition <= 10 {
			table1.SetSelectable(true, false)
			table2.SetSelectable(false, false)
			table1.Select(globalPosition, 0)
			app.SetFocus(table1)
		} else {
			table1.SetSelectable(false, false)
			table2.SetSelectable(true, false)
			table2.Select(globalPosition-10, 0)
			app.SetFocus(table2)
		}
	}

	updateSelection()

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' || event.Key() == tcell.KeyEscape {
			app.Stop()
			return nil
		}

		switch event.Key() {
		case tcell.KeyUp:
			globalPosition--
			if globalPosition < 1 {
				globalPosition = 20
			}
			updateSelection()
			return nil
		case tcell.KeyDown:
			globalPosition++
			if globalPosition > 20 {
				globalPosition = 1
			}
			updateSelection()
			return nil
		case tcell.KeyLeft:
			if globalPosition > 10 {
				globalPosition -= 10
			}
			updateSelection()
			return nil
		case tcell.KeyRight:
			if globalPosition <= 10 {
				globalPosition += 10
			}
			updateSelection()
			return nil
		}

		return event
	})

	app.SetRoot(flex, true).SetFocus(table1)
}
