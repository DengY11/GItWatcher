package cmd

import (
	"encoding/json"
	"fmt"

	"git-watcher/pkg/analyzer"
	"git-watcher/pkg/scanner"
	"git-watcher/pkg/stats"

	"github.com/spf13/cobra"
)

var (
	rootPath string
	output   string
)

var rootCmd = &cobra.Command{
	Use:   "git-watcher",
	Short: "Git repository statistics tool",
	Long:  `Recursively scan the specified directory for Git repositories and compute statistics such as authors, commit times, late-night commits, etc.`,
	RunE:  run,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&rootPath, "path", "p", ".", "Directory path to scan")
	rootCmd.Flags().StringVarP(&output, "output", "o", "json", "Output format (json|text)")
}

func run(cmd *cobra.Command, args []string) error {
	gitScanner := scanner.NewGitScanner()
	repos, err := gitScanner.ScanDirectory(rootPath)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(repos) == 0 {
		fmt.Println("No Git repositories found")
		return nil
	}

	allStats := make(map[string]interface{})

	for _, repo := range repos {
		fmt.Printf("Analyzing repository: %s\n", repo)
		fmt.Println("Large repositories may take time")

		analyzer := analyzer.NewGitAnalyzer(repo)
		commits, err := analyzer.GetCommitInfo()
		if err != nil {
			fmt.Printf("Failed to analyze repository %s: %v\n", repo, err)
			continue
		}

		calculator := stats.NewStatsCalculator()
		repoStats := calculator.CalculateAll(commits)

		allStats[repo] = map[string]interface{}{
			"total_commits": len(commits),
			"statistics":    repoStats,
		}
	}

	switch output {
	case "json":
		jsonOutput, err := json.MarshalIndent(allStats, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to generate JSON output: %w", err)
		}
		fmt.Println(string(jsonOutput))
	case "text":
		printTextOutput(allStats)
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}

	return nil
}

func printTextOutput(allStats map[string]interface{}) {
	for repo, repoData := range allStats {
		fmt.Printf("\n=== Repository: %s ===\n", repo)

		data := repoData.(map[string]interface{})
		fmt.Printf("Total commits: %v\n", data["total_commits"])

		stats := data["statistics"].(map[string]interface{})

		if latestCommit := stats["latest_commit"]; latestCommit != nil {
			commit := latestCommit.(analyzer.CommitInfo)
			fmt.Printf("Latest commit: %s by %s at %s\n",
				commit.Hash[:7], commit.Author, commit.Date.Format("2006-01-02 15:04:05"))
		}

		if authorCounts := stats["commit_count_by_author"]; authorCounts != nil {
			fmt.Println("\nAuthor statistics:")
			authors := authorCounts.(map[string]int)
			for author, count := range authors {
				fmt.Printf("  %s: %d\n", author, count)
			}
		}

		if lateNight := stats["late_night_commits"]; lateNight != nil {
			lateNightData := lateNight.(map[string]interface{})
			fmt.Printf("\nLate-night commits (23:00-06:00): %v\n", lateNightData["total"])
			if authors := lateNightData["authors"].(map[string]int); len(authors) > 0 {
				fmt.Println("Late-night authors:")
				for author, count := range authors {
					fmt.Printf("  %s: %d\n", author, count)
				}
			}
		}

		if weekend := stats["weekend_commits"]; weekend != nil {
			weekendData := weekend.(map[string]interface{})
			fmt.Printf("\nWeekend commits: %v\n", weekendData["total"])
			if authors := weekendData["authors"].(map[string]int); len(authors) > 0 {
				fmt.Println("Weekend authors:")
				for author, count := range authors {
					fmt.Printf("  %s: %d\n", author, count)
				}
			}
		}
		if lineCountByAuthor := stats["commit_line_count_by_author"]; lineCountByAuthor != nil {
			fmt.Println("\nLines changed by author:")
			lineCounts := lineCountByAuthor.(map[string]int64)
			for author, count := range lineCounts {
				fmt.Printf("  %s: %d\n", author, count)
			}
		}
	}
}
