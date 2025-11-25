package tui

import (
	"fmt"
	"strings"
	"time"

	"git-watcher/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func StartTUI(rootPath string) error {
	app := tview.NewApplication()
	ctrl := ui.NewController(rootPath)

	header := tview.NewTextView().SetDynamicColors(true)
	header.SetText("[yellow]GitWatcher TUI [-] [blue](q)quit (r)refresh (a)auto-refresh (e)export JSON (1)total commits (2)commits (3)authors (4)timeline (f)author filter")

	repos := tview.NewList()
	overview := tview.NewTextView().SetDynamicColors(true)
	commits := tview.NewTextView().SetDynamicColors(true)
	authors := tview.NewTextView().SetDynamicColors(true)
	timeline := tview.NewTextView().SetDynamicColors(true)

	right := tview.NewPages()
	right.AddPage("overview", overview, true, true)
	right.AddPage("commits", commits, true, false)
	right.AddPage("authors", authors, true, false)
	right.AddPage("timeline", timeline, true, false)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(tview.NewFlex().AddItem(repos, 30, 0, true).AddItem(right, 0, 1, false), 0, 1, true)

	selectedRepo := ""
	authorFilter := ""
	auto := false
	var ticker *time.Ticker

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

	renderCommits := func() {
		b := &strings.Builder{}
		if selectedRepo == "" {
			fmt.Fprintln(b, "No repository selected")
		} else {
			for _, c := range ctrl.State.CommitsByRepo[selectedRepo] {
				if authorFilter != "" && c.Author != authorFilter {
					continue
				}
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
				for a, v := range m {
					fmt.Fprintf(b, "%s: %d\n", a, v)
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
				for k := 0; k < 24; k++ {
					fmt.Fprintf(b, "%02d: %d\n", k, m[k])
				}
			}
		}
		timeline.SetText(b.String())
	}

	refresh := func() {
		go func() {
			_ = ctrl.Refresh()
			app.QueueUpdateDraw(func() {
				repos.Clear()
				for _, r := range ctrl.State.Repos {
					repos.AddItem(r, "", 0, nil)
				}
				if selectedRepo == "" && len(ctrl.State.Repos) > 0 {
					selectedRepo = ctrl.State.Repos[0]
				}
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
	})

	layout.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Rune() {
		case 'q':
			app.Stop()
		case 'r':
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
			data, _ := ctrl.ExportJSON()
			modal := tview.NewModal().SetText(string(data)).AddButtons([]string{"Close"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) { right.RemovePage("modal"); app.Draw() })
			right.AddAndSwitchToPage("modal", modal, true)
		case 'f':
			input := tview.NewInputField().SetLabel("Author filter:")
			form := tview.NewForm().AddFormItem(input).AddButton("OK", func() { authorFilter = input.GetText(); right.RemovePage("filter"); renderCommits(); app.Draw() }).AddButton("Cancel", func() { right.RemovePage("filter"); app.Draw() })
			right.AddAndSwitchToPage("filter", form, true)
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
