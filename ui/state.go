package ui

import (
    "encoding/json"
    "git-watcher/pkg/analyzer"
    "git-watcher/pkg/scanner"
    "git-watcher/pkg/stats"
)

type AppState struct {
    RootPath      string
    Repos         []string
    CommitsByRepo map[string][]analyzer.CommitInfo
    StatsByRepo   map[string]map[string]interface{}
    Loading       bool
}

type Controller struct {
    State *AppState
}

func NewController(root string) *Controller {
    return &Controller{State: &AppState{
        RootPath:      root,
        Repos:         []string{},
        CommitsByRepo: map[string][]analyzer.CommitInfo{},
        StatsByRepo:   map[string]map[string]interface{}{},
        Loading:       false,
    }}
}

func (c *Controller) Refresh() error {
    c.State.Loading = true
    gitScanner := scanner.NewGitScanner()
    repos, err := gitScanner.ScanDirectory(c.State.RootPath)
    if err != nil {
        c.State.Loading = false
        return err
    }
    c.State.Repos = repos
    c.State.CommitsByRepo = map[string][]analyzer.CommitInfo{}
    c.State.StatsByRepo = map[string]map[string]interface{}{}
    for _, repo := range repos {
        a := analyzer.NewGitAnalyzer(repo)
        commits, err := a.GetCommitInfo()
        if err != nil {
            continue
        }
        c.State.CommitsByRepo[repo] = commits
        calc := stats.NewStatsCalculator()
        c.State.StatsByRepo[repo] = calc.CalculateAll(commits)
    }
    c.State.Loading = false
    return nil
}

func (c *Controller) ExportJSON() ([]byte, error) {
    out := map[string]interface{}{}
    for repo, commits := range c.State.CommitsByRepo {
        out[repo] = map[string]interface{}{
            "total_commits": len(commits),
            "statistics":    c.State.StatsByRepo[repo],
        }
    }
    return json.MarshalIndent(out, "", "  ")
}
