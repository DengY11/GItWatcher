package analyzer

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type CommitInfo struct {
	Author    string
	Email     string
	Date      time.Time
	Message   string
	Hash      string
	LineCount int64
}

type GitAnalyzer struct {
	repoPath string
}

func NewGitAnalyzer(repoPath string) *GitAnalyzer {
	return &GitAnalyzer{repoPath: repoPath}
}

func (ga *GitAnalyzer) GetCommitInfo() ([]CommitInfo, error) {
	repo, err := git.PlainOpen(ga.repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	var commitHashes []plumbing.Hash
	err = commitIter.ForEach(func(c *object.Commit) error {
		commitHashes = append(commitHashes, c.Hash)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits for hashes: %w", err)
	}

	numWorkers := runtime.NumCPU()
	hashChan := make(chan plumbing.Hash, len(commitHashes))
	resultsChan := make(chan CommitInfo, len(commitHashes))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			workerRepo, err := git.PlainOpen(ga.repoPath)
			if err != nil {
				return
			}

			for hash := range hashChan {
				c, err := workerRepo.CommitObject(hash)
				if err != nil {
					continue
				}

				stats, err := c.Stats()
				if err != nil {
					continue
				}
				var totalLines int64
				for _, stat := range stats {
					totalLines += int64(stat.Addition + stat.Deletion)
				}

				resultsChan <- CommitInfo{
					Author:    c.Author.Name,
					Email:     c.Author.Email,
					Date:      c.Author.When,
					Message:   c.Message,
					Hash:      c.Hash.String(),
					LineCount: totalLines,
				}
			}
		}()
	}

	for _, hash := range commitHashes {
		hashChan <- hash
	}
	close(hashChan)

	wg.Wait()
	close(resultsChan)

	var commits []CommitInfo
	for commitInfo := range resultsChan {
		commits = append(commits, commitInfo)
	}

	return commits, nil
}

func (ga *GitAnalyzer) GetCommitInfoWithProgress(onProgress func(processed int, total int)) ([]CommitInfo, error) {
	repo, err := git.PlainOpen(ga.repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	var commitHashes []plumbing.Hash
	err = commitIter.ForEach(func(c *object.Commit) error {
		commitHashes = append(commitHashes, c.Hash)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits for hashes: %w", err)
	}

	total := len(commitHashes)

	numWorkers := runtime.NumCPU()
	hashChan := make(chan plumbing.Hash, len(commitHashes))
	resultsChan := make(chan CommitInfo, len(commitHashes))
	var wg sync.WaitGroup
	processed := 0

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			workerRepo, err := git.PlainOpen(ga.repoPath)
			if err != nil {
				return
			}

			for hash := range hashChan {
				c, err := workerRepo.CommitObject(hash)
				if err != nil {
					processed++
					if onProgress != nil {
						onProgress(processed, total)
					}
					continue
				}

				stats, err := c.Stats()
				if err != nil {
					processed++
					if onProgress != nil {
						onProgress(processed, total)
					}
					continue
				}
				var totalLines int64
				for _, stat := range stats {
					totalLines += int64(stat.Addition + stat.Deletion)
				}

				resultsChan <- CommitInfo{
					Author:    c.Author.Name,
					Email:     c.Author.Email,
					Date:      c.Author.When,
					Message:   c.Message,
					Hash:      c.Hash.String(),
					LineCount: totalLines,
				}
				processed++
				if onProgress != nil {
					onProgress(processed, total)
				}
			}
		}()
	}

	for _, hash := range commitHashes {
		hashChan <- hash
	}
	close(hashChan)

	wg.Wait()
	close(resultsChan)

	var commits []CommitInfo
	for commitInfo := range resultsChan {
		commits = append(commits, commitInfo)
	}

	return commits, nil
}
