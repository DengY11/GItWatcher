package analyzer

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type CommitInfo struct {
	Author    string
	Email     string
	Date      time.Time
	Message   string
	Hash      string
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

	var commits []CommitInfo
	err = commitIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, CommitInfo{
			Author:  c.Author.Name,
			Email:   c.Author.Email,
			Date:    c.Author.When,
			Message: c.Message,
			Hash:    c.Hash.String(),
		})
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, nil
}