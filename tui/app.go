package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"git-watcher/pkg/analyzer"
	"git-watcher/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func StartTUI(rootPath string) error {
	app := tview.NewApplication()
	ctrl := ui.NewController(rootPath)

	repos := tview.NewList()
	overview := tview.NewTextView().SetDynamicColors(true)
	overview.SetScrollable(true)
	commits := tview.NewTextView().SetDynamicColors(true)
	commits.SetScrollable(true)
	authors := tview.NewTextView().SetDynamicColors(true)
	authors.SetScrollable(true)
	timeline := tview.NewTextView().SetDynamicColors(true)
	timeline.SetScrollable(true)

	statusView := tview.NewTextView().SetDynamicColors(true)
	statusView.SetBorder(true)
	helpBar := tview.NewFlex().SetDirection(tview.FlexColumn)
	mk := func(text string) *tview.TextView {
		tv := tview.NewTextView()
		tv.SetText(text)
		tv.SetBorder(true)
		return tv
	}
	helpBar.AddItem(mk("q Quit"), 0, 1, false)
	helpBar.AddItem(mk("r Focus repos"), 0, 1, false)
	helpBar.AddItem(mk("R Refresh"), 0, 1, false)
	helpBar.AddItem(mk("Enter Select"), 0, 1, false)
	helpBar.AddItem(mk("Up/Down Scroll"), 0, 1, false)
	helpBar.AddItem(mk("a Auto-refresh"), 0, 1, false)
	helpBar.AddItem(mk("e Export JSON"), 0, 1, false)
	// removed author filter entry
	helpBar.AddItem(mk("j/k Select repo"), 0, 1, false)
	helpBar.AddItem(mk("s Toggle sort"), 0, 1, false)
	helpBar.AddItem(mk("1 Overview"), 0, 1, false)
	helpBar.AddItem(mk("2 Commits"), 0, 1, false)
	helpBar.AddItem(mk("3 Authors"), 0, 1, false)
	helpBar.AddItem(mk("4 Timeline"), 0, 1, false)

	right := tview.NewPages()
	right.SetBorder(true)
	right.AddPage("overview", overview, true, true)
	right.AddPage("commits", commits, true, false)
	right.AddPage("authors", authors, true, false)
	right.AddPage("timeline", timeline, true, false)

	content := tview.NewFlex().AddItem(repos, 30, 0, true).AddItem(right, 0, 1, false)
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(content, 0, 1, true).
		AddItem(statusView, 3, 0, false).
		AddItem(helpBar, 3, 0, false)

	selectedRepo := ""
	focusOnRepos := false
	scrollOverviewY := 0
	scrollCommitsY := 0
	scrollAuthorsY := 0
	scrollTimelineY := 0
	// removed author filter
	auto := false
	var ticker *time.Ticker
	dialogOpen := false

	renderOverview := func() {
		b := &strings.Builder{}
		if selectedRepo == "" {
			fmt.Fprintln(b, "No repository selected")
		} else {
			stats := ctrl.State.StatsByRepo[selectedRepo]
			commitsList := ctrl.State.CommitsByRepo[selectedRepo]
			fmt.Fprintf(b, "Repo: %s\n", selectedRepo)
			fmt.Fprintf(b, "Total commits: %d\n", len(commitsList))
			if v := stats["late_night_commits"]; v != nil {
				m := v.(map[string]interface{})
				fmt.Fprintf(b, "Late-night: %v\n", m["total"])
			}
			if v := stats["weekend_commits"]; v != nil {
				m := v.(map[string]interface{})
				fmt.Fprintf(b, "Weekend: %v\n", m["total"])
			}
			if v := stats["commit_activity_by_hour"]; v != nil {
				m := v.(map[int]int)
				fmt.Fprint(b, "Hourly: ")
				for h := 0; h < 24; h++ {
					fmt.Fprintf(b, "%02d=%d ", h, m[h])
				}
				fmt.Fprintln(b)
			}
		}
		overview.SetText(b.String())
	}

	sortAscCommits := false
	sortAscAuthors := false
	sortAscTimeline := true

	renderCommits := func() {
		b := &strings.Builder{}
		if selectedRepo == "" {
			fmt.Fprintln(b, "No repository selected")
		} else {
			list := append([]analyzer.CommitInfo(nil), ctrl.State.CommitsByRepo[selectedRepo]...)
			if sortAscCommits {
				sort.Slice(list, func(i, j int) bool { return list[i].Date.Before(list[j].Date) })
			} else {
				sort.Slice(list, func(i, j int) bool { return list[i].Date.After(list[j].Date) })
			}
			for _, c := range list {
				fmt.Fprintf(b, "%s %s %s\n", c.Hash[:7], c.Author, c.Date.Format("2006-01-02 15:04:05"))
				fmt.Fprintln(b, c.Message)
			}
		}
		commits.SetText(b.String())
	}

	renderAuthors := func() {
		b := &strings.Builder{}
		if selectedRepo == "" {
			fmt.Fprintln(b, "No repository selected")
		} else {
			stat := ctrl.State.StatsByRepo[selectedRepo]["commit_count_by_author"]
			if stat != nil {
				m := stat.(map[string]int)
				type kv struct {
					A string
					V int
				}
				arr := make([]kv, 0, len(m))
				for a, v := range m {
					arr = append(arr, kv{a, v})
				}
				if sortAscAuthors {
					sort.Slice(arr, func(i, j int) bool { return arr[i].V < arr[j].V })
				} else {
					sort.Slice(arr, func(i, j int) bool { return arr[i].V > arr[j].V })
				}
				for _, it := range arr {
					fmt.Fprintf(b, "%s: %d\n", it.A, it.V)
				}
			}
		}
		authors.SetText(b.String())
	}

	renderTimeline := func() {
		b := &strings.Builder{}
		if selectedRepo == "" {
			fmt.Fprintln(b, "No repository selected")
		} else {
			h := ctrl.State.StatsByRepo[selectedRepo]["commit_activity_by_hour"]
			if h != nil {
				m := h.(map[int]int)
				max := 0
				for k := 0; k < 24; k++ {
					if m[k] > max {
						max = m[k]
					}
				}
				width := 40
				drawBar := func(val int) string {
					if max == 0 {
						return strings.Repeat(".", width)
					}
					filled := val * width / max
					if filled < 0 {
						filled = 0
					}
					if filled > width {
						filled = width
					}
					return strings.Repeat("█", filled) + strings.Repeat(" ", width-filled)
				}
				if sortAscTimeline {
					for k := 0; k < 24; k++ {
						fmt.Fprintf(b, "%02d: %s %d\n", k, drawBar(m[k]), m[k])
					}
				} else {
					for k := 23; k >= 0; k-- {
						fmt.Fprintf(b, "%02d: %s %d\n", k, drawBar(m[k]), m[k])
					}
				}
			}
		}
		timeline.SetText(b.String())
	}

	refresh := func() {
		statusView.SetText("Analyzing...")
		go func() {
			_ = ctrl.RefreshWithProgress(func(p ui.Progress) {
				app.QueueUpdateDraw(func() {
					barLen := 20
					pct := 0
					if p.Total > 0 {
						pct = int(float64(p.Processed) / float64(p.Total) * 100)
					}
					filled := pct * barLen / 100
					bar := strings.Repeat("#", filled) + strings.Repeat(".", barLen-filled)
					statusView.SetText(fmt.Sprintf("Analyzing %s [%s] %d%% (%d/%d)", p.Repo, bar, pct, p.Processed, p.Total))
				})
			})
			app.QueueUpdateDraw(func() {
				repos.Clear()
				for _, r := range ctrl.State.Repos {
					repos.AddItem(r, "", 0, nil)
				}
				if selectedRepo == "" && len(ctrl.State.Repos) > 0 {
					selectedRepo = ctrl.State.Repos[0]
				}
				statusView.SetText("Idle")
				focusOnRepos = false
				app.SetFocus(right)
				right.SetBorderColor(tcell.ColorYellow)
				repos.SetBorder(true)
				repos.SetBorderColor(tcell.ColorGray)
				renderOverview()
				renderCommits()
				renderAuthors()
				renderTimeline()
			})
		}()
	}

	repos.SetSelectedFunc(func(i int, mainText, secondaryText string, shortcut rune) {
		selectedRepo = mainText
		renderOverview()
		renderCommits()
		renderAuthors()
		renderTimeline()
		focusOnRepos = false
		app.SetFocus(right)
		right.SetBorderColor(tcell.ColorYellow)
		repos.SetBorder(true)
		repos.SetBorderColor(tcell.ColorGray)
	})

	repos.SetChangedFunc(func(i int, mainText, secondaryText string, shortcut rune) {
		selectedRepo = mainText
		scrollOverviewY = 0
		scrollCommitsY = 0
		scrollAuthorsY = 0
		scrollTimelineY = 0
		renderOverview()
		renderCommits()
		renderAuthors()
		renderTimeline()
	})

	layout.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if dialogOpen {
			return ev
		}
		switch ev.Key() {
		case tcell.KeyUp:
			if !focusOnRepos {
				name, _ := right.GetFrontPage()
				switch name {
				case "overview":
					if scrollOverviewY > 0 {
						scrollOverviewY--
					}
					overview.ScrollTo(0, scrollOverviewY)
				case "commits":
					if scrollCommitsY > 0 {
						scrollCommitsY--
					}
					commits.ScrollTo(0, scrollCommitsY)
				case "authors":
					if scrollAuthorsY > 0 {
						scrollAuthorsY--
					}
					authors.ScrollTo(0, scrollAuthorsY)
				case "timeline":
					if scrollTimelineY > 0 {
						scrollTimelineY--
					}
					timeline.ScrollTo(0, scrollTimelineY)
				}
				return nil
			}
		case tcell.KeyDown:
			if !focusOnRepos {
				name, _ := right.GetFrontPage()
				switch name {
				case "overview":
					scrollOverviewY++
					overview.ScrollTo(0, scrollOverviewY)
				case "commits":
					scrollCommitsY++
					commits.ScrollTo(0, scrollCommitsY)
				case "authors":
					scrollAuthorsY++
					authors.ScrollTo(0, scrollAuthorsY)
				case "timeline":
					scrollTimelineY++
					timeline.ScrollTo(0, scrollTimelineY)
				}
				return nil
			}
		case tcell.KeyTab, tcell.KeyEsc:
			if focusOnRepos {
				focusOnRepos = false
				name, _ := right.GetFrontPage()
				switch name {
				case "overview":
					app.SetFocus(overview)
				case "commits":
					app.SetFocus(commits)
				case "authors":
					app.SetFocus(authors)
				case "timeline":
					app.SetFocus(timeline)
				}
				right.SetBorderColor(tcell.ColorYellow)
				repos.SetBorder(true)
				repos.SetBorderColor(tcell.ColorGray)
				return nil
			}
		}
		switch ev.Rune() {
		case 'q':
			app.Stop()
		case 'r':
			if focusOnRepos {
				focusOnRepos = false
				name, _ := right.GetFrontPage()
				switch name {
				case "overview":
					app.SetFocus(overview)
				case "commits":
					app.SetFocus(commits)
				case "authors":
					app.SetFocus(authors)
				case "timeline":
					app.SetFocus(timeline)
				}
				right.SetBorderColor(tcell.ColorYellow)
				repos.SetBorder(true)
				repos.SetBorderColor(tcell.ColorGray)
			} else {
				focusOnRepos = true
				app.SetFocus(repos)
				repos.SetBorder(true)
				repos.SetBorderColor(tcell.ColorYellow)
				right.SetBorderColor(tcell.ColorGray)
			}
			return nil
		case 'R':
			refresh()
		case '1':
			right.SwitchToPage("overview")
		case '2':
			right.SwitchToPage("commits")
		case '3':
			right.SwitchToPage("authors")
		case '4':
			right.SwitchToPage("timeline")
		case 'e':
			defaultPath := filepath.Join(ctrl.State.RootPath, "gitwatcher.json")
			input := tview.NewInputField().SetLabel("Save path:").SetText(defaultPath)
			form := tview.NewForm().AddFormItem(input)
			doSave := func() {
				data, err := ctrl.ExportJSON()
				if err != nil {
					modal := tview.NewModal().SetText(fmt.Sprintf("Export failed: %v", err)).AddButtons([]string{"OK"}).SetDoneFunc(func(i int, l string) { dialogOpen = false; app.SetRoot(layout, true) })
					dialogOpen = true
					app.SetRoot(modal, true)
					return
				}
				path := strings.TrimSpace(input.GetText())
				if path == "" {
					path = defaultPath
				}
				if err := os.WriteFile(path, data, 0644); err != nil {
					modal := tview.NewModal().SetText(fmt.Sprintf("Write failed: %v", err)).AddButtons([]string{"OK"}).SetDoneFunc(func(i int, l string) { dialogOpen = false; app.SetRoot(layout, true) })
					dialogOpen = true
					app.SetRoot(modal, true)
					return
				}
				statusView.SetText(fmt.Sprintf("Exported: %s", path))
				modal := tview.NewModal().SetText(fmt.Sprintf("Saved to %s", path)).AddButtons([]string{"OK"}).SetDoneFunc(func(i int, l string) { dialogOpen = false; app.SetRoot(layout, true) })
				dialogOpen = true
				app.SetRoot(modal, true)
			}
			form.AddButton("Save", doSave)
			form.AddButton("Cancel", func() { dialogOpen = false; app.SetRoot(layout, true) })
			form.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
				if ev.Key() == tcell.KeyEnter {
					doSave()
					return nil
				}
				return ev
			})
			dialogOpen = true
			app.SetRoot(form, true)
			// removed author filter
		case 's':
			name, _ := right.GetFrontPage()
			switch name {
			case "commits":
				sortAscCommits = !sortAscCommits
				renderCommits()
			case "authors":
				sortAscAuthors = !sortAscAuthors
				renderAuthors()
			case "timeline":
				sortAscTimeline = !sortAscTimeline
				renderTimeline()
			}
		case 'j':
			if focusOnRepos {
				idx := repos.GetCurrentItem()
				if idx+1 < len(ctrl.State.Repos) {
					repos.SetCurrentItem(idx + 1)
					selectedRepo = ctrl.State.Repos[idx+1]
					renderOverview()
					renderCommits()
					renderAuthors()
					renderTimeline()
				}
				return nil
			} else {
				name, _ := right.GetFrontPage()
				switch name {
				case "overview":
					scrollOverviewY++
					overview.ScrollTo(0, scrollOverviewY)
				case "commits":
					scrollCommitsY++
					commits.ScrollTo(0, scrollCommitsY)
				case "authors":
					scrollAuthorsY++
					authors.ScrollTo(0, scrollAuthorsY)
				case "timeline":
					scrollTimelineY++
					timeline.ScrollTo(0, scrollTimelineY)
				}
				return nil
			}
		case 'k':
			if focusOnRepos {
				idx := repos.GetCurrentItem()
				if idx-1 >= 0 {
					repos.SetCurrentItem(idx - 1)
					selectedRepo = ctrl.State.Repos[idx-1]
					renderOverview()
					renderCommits()
					renderAuthors()
					renderTimeline()
				}
				return nil
			} else {
				name, _ := right.GetFrontPage()
				switch name {
				case "overview":
					if scrollOverviewY > 0 {
						scrollOverviewY--
					}
					overview.ScrollTo(0, scrollOverviewY)
				case "commits":
					if scrollCommitsY > 0 {
						scrollCommitsY--
					}
					commits.ScrollTo(0, scrollCommitsY)
				case "authors":
					if scrollAuthorsY > 0 {
						scrollAuthorsY--
					}
					authors.ScrollTo(0, scrollAuthorsY)
				case "timeline":
					if scrollTimelineY > 0 {
						scrollTimelineY--
					}
					timeline.ScrollTo(0, scrollTimelineY)
				}
				return nil
			}
		case 'a':
			if !auto {
				auto = true
				ticker = time.NewTicker(30 * time.Second)
				go func() {
					for range ticker.C {
						refresh()
					}
				}()
			} else {
				auto = false
				if ticker != nil {
					ticker.Stop()
				}
			}
		}
		return ev
	})

	go refresh()
	if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
		return err
	}
	return nil
}
