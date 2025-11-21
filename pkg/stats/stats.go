package stats

import (
	"git-watcher/pkg/analyzer"
	"time"
)

type Statistics interface {
	Calculate(commits []analyzer.CommitInfo) interface{}
	Name() string
}

type CommitCountByAuthor struct{}

func (c *CommitCountByAuthor) Name() string {
	return "commit_count_by_author"
}

func (c *CommitCountByAuthor) Calculate(commits []analyzer.CommitInfo) interface{} {
	authorCount := make(map[string]int)
	for _, commit := range commits {
		authorCount[commit.Author]++
	}
	return authorCount
}

type LatestCommit struct{}

func (l *LatestCommit) Name() string {
	return "latest_commit"
}

func (l *LatestCommit) Calculate(commits []analyzer.CommitInfo) interface{} {
	if len(commits) == 0 {
		return nil
	}

	latest := commits[0]
	for _, commit := range commits {
		if commit.Date.After(latest.Date) {
			latest = commit
		}
	}
	return latest
}

type LateNightCommits struct{}

func (l *LateNightCommits) Name() string {
	return "late_night_commits"
}

func (l *LateNightCommits) Calculate(commits []analyzer.CommitInfo) interface{} {
	lateNightCount := 0
	lateNightAuthors := make(map[string]int)

	for _, commit := range commits {
		hour := commit.Date.Hour()
		if hour >= 23 || hour <= 6 {
			lateNightCount++
			lateNightAuthors[commit.Author]++
		}
	}

	return map[string]interface{}{
		"total":   lateNightCount,
		"authors": lateNightAuthors,
	}
}

type CommitActivityByHour struct{}

func (c *CommitActivityByHour) Name() string {
	return "commit_activity_by_hour"
}

func (c *CommitActivityByHour) Calculate(commits []analyzer.CommitInfo) interface{} {
	hourlyActivity := make(map[int]int)
	for _, commit := range commits {
		hour := commit.Date.Hour()
		hourlyActivity[hour]++
	}
	return hourlyActivity
}

type StatsCalculator struct {
	statistics []Statistics
}

func NewStatsCalculator() *StatsCalculator {
	return &StatsCalculator{
		statistics: []Statistics{
			&CommitCountByAuthor{},
			&LatestCommit{},
			&LateNightCommits{},
			&CommitActivityByHour{},
			&WeekendCommits{},
		},
	}
}

func (sc *StatsCalculator) AddStatistic(stat Statistics) {
	sc.statistics = append(sc.statistics, stat)
}

func (sc *StatsCalculator) CalculateAll(commits []analyzer.CommitInfo) map[string]interface{} {
	results := make(map[string]interface{})

	for _, stat := range sc.statistics {
		results[stat.Name()] = stat.Calculate(commits)
	}

	return results
}

type WeekendCommits struct{
	//this is fucking truly work life balance
}

func (w *WeekendCommits) Name() string {
	return "weekend_commits"
}

func (w *WeekendCommits) Calculate(commits []analyzer.CommitInfo) interface{} {
	weekendCount := 0
	weekendAuthors := make(map[string]int)

	isWeekend := func(date time.Time) bool {
		weekday := date.Weekday()
		return weekday == time.Saturday || weekday == time.Sunday
	}
	for _, commit := range commits {
		if isWeekend(commit.Date) {
			weekendCount++
			weekendAuthors[commit.Author]++
		}
	}

	return map[string]interface{}{
		"total":   weekendCount,
		"authors": weekendAuthors,
	}
}
