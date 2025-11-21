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
	Short: "Git仓库统计工具",
	Long:  `递归扫描指定目录下的所有Git仓库，并统计各种信息如提交者、提交时间、深夜提交次数等`,
	RunE:  run,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&rootPath, "path", "p", ".", "要扫描的目录路径")
	rootCmd.Flags().StringVarP(&output, "output", "o", "json", "输出格式 (json|text)")
}

func run(cmd *cobra.Command, args []string) error {
	gitScanner := scanner.NewGitScanner()
	repos, err := gitScanner.ScanDirectory(rootPath)
	if err != nil {
		return fmt.Errorf("扫描目录失败: %w", err)
	}

	if len(repos) == 0 {
		fmt.Println("未找到Git仓库")
		return nil
	}

	allStats := make(map[string]interface{})

	for _, repo := range repos {
		fmt.Printf("正在分析仓库: %s\n", repo)

		analyzer := analyzer.NewGitAnalyzer(repo)
		commits, err := analyzer.GetCommitInfo()
		if err != nil {
			fmt.Printf("分析仓库 %s 失败: %v\n", repo, err)
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
			return fmt.Errorf("生成JSON输出失败: %w", err)
		}
		fmt.Println(string(jsonOutput))
	case "text":
		printTextOutput(allStats)
	default:
		return fmt.Errorf("不支持的输出格式: %s", output)
	}

	return nil
}

func printTextOutput(allStats map[string]interface{}) {
	for repo, repoData := range allStats {
		fmt.Printf("\n=== 仓库: %s ===\n", repo)

		data := repoData.(map[string]interface{})
		fmt.Printf("总提交数: %v\n", data["total_commits"])

		stats := data["statistics"].(map[string]interface{})

		if latestCommit := stats["latest_commit"]; latestCommit != nil {
			commit := latestCommit.(analyzer.CommitInfo)
			fmt.Printf("最新提交: %s by %s at %s\n",
				commit.Hash[:7], commit.Author, commit.Date.Format("2006-01-02 15:04:05"))
		}

		if authorCounts := stats["commit_count_by_author"]; authorCounts != nil {
			fmt.Println("\n提交者统计:")
			authors := authorCounts.(map[string]int)
			for author, count := range authors {
				fmt.Printf("  %s: %d次\n", author, count)
			}
		}

		if lateNight := stats["late_night_commits"]; lateNight != nil {
			lateNightData := lateNight.(map[string]interface{})
			fmt.Printf("\n深夜提交 (23:00-06:00): %v次\n", lateNightData["total"])
			if authors := lateNightData["authors"].(map[string]int); len(authors) > 0 {
				fmt.Println("深夜提交者:")
				for author, count := range authors {
					fmt.Printf("  %s: %d次\n", author, count)
				}
			}
		}

		if weekend := stats["weekend_commits"]; weekend != nil {
			weekendData := weekend.(map[string]interface{})
			fmt.Printf("\n周末提交: %v次\n", weekendData["total"])
			if authors := weekendData["authors"].(map[string]int); len(authors) > 0 {
				fmt.Println("周末提交者:")
				for author, count := range authors {
					fmt.Printf("  %s: %d次\n", author, count)
				}
			}
		}

	}
}
