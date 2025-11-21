package stats

import (
	"testing"
	"time"

	"git-watcher/pkg/analyzer"
)

func TestCommitCountByAuthor(t *testing.T) {
	commits := []analyzer.CommitInfo{
		{Author: "Alice", Date: time.Now()},
		{Author: "Bob", Date: time.Now()},
		{Author: "Alice", Date: time.Now()},
	}

	stat := &CommitCountByAuthor{}
	result := stat.Calculate(commits).(map[string]int)

	if result["Alice"] != 2 {
		t.Errorf("Expected Alice to have 2 commits, got %d", result["Alice"])
	}
	if result["Bob"] != 1 {
		t.Errorf("Expected Bob to have 1 commit, got %d", result["Bob"])
	}
}

func TestLatestCommit(t *testing.T) {
	now := time.Now()
	commits := []analyzer.CommitInfo{
		{Author: "Alice", Date: now.Add(-1 * time.Hour)},
		{Author: "Bob", Date: now},
		{Author: "Charlie", Date: now.Add(-2 * time.Hour)},
	}

	stat := &LatestCommit{}
	result := stat.Calculate(commits).(analyzer.CommitInfo)

	if result.Author != "Bob" {
		t.Errorf("Expected latest commit to be by Bob, got %s", result.Author)
	}
}

func TestLateNightCommits(t *testing.T) {
	commits := []analyzer.CommitInfo{
		{Author: "Alice", Date: time.Date(2024, 1, 1, 23, 30, 0, 0, time.UTC)}, // 23:30 - late night
		{Author: "Bob", Date: time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC)},    // 02:00 - late night
		{Author: "Charlie", Date: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)}, // 14:00 - not late night
	}

	stat := &LateNightCommits{}
	result := stat.Calculate(commits).(map[string]interface{})

	total := result["total"].(int)
	if total != 2 {
		t.Errorf("Expected 2 late night commits, got %d", total)
	}

	authors := result["authors"].(map[string]int)
	if authors["Alice"] != 1 {
		t.Errorf("Expected Alice to have 1 late night commit, got %d", authors["Alice"])
	}
	if authors["Bob"] != 1 {
		t.Errorf("Expected Bob to have 1 late night commit, got %d", authors["Bob"])
	}
}

func TestWeekendCommits(t *testing.T) {
	commits := []analyzer.CommitInfo{
		{Author: "Alice", Date: time.Date(2024, 7, 27, 10, 0, 0, 0, time.UTC)}, // Saturday
		{Author: "Bob", Date: time.Date(2024, 7, 28, 11, 0, 0, 0, time.UTC)},   // Sunday
		{Author: "Alice", Date: time.Date(2024, 7, 28, 12, 0, 0, 0, time.UTC)},  // Sunday
		{Author: "Charlie", Date: time.Date(2024, 7, 29, 9, 0, 0, 0, time.UTC)}, // Monday
	}

	stat := &WeekendCommits{}
	result := stat.Calculate(commits).(map[string]interface{})

	total := result["total"].(int)
	if total != 3 {
		t.Errorf("Expected 3 weekend commits, got %d", total)
	}

	authors := result["authors"].(map[string]int)
	if authors["Alice"] != 2 {
		t.Errorf("Expected Alice to have 2 weekend commits, got %d", authors["Alice"])
	}
	if authors["Bob"] != 1 {
		t.Errorf("Expected Bob to have 1 weekend commit, got %d", authors["Bob"])
	}
	if _, ok := authors["Charlie"]; ok {
		t.Errorf("Expected Charlie to have 0 weekend commits, but was found in authors map")
	}
}