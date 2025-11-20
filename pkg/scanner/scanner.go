package scanner

import (
	"os"
	"path/filepath"
)

type GitScanner struct {
	gitRepos []string
}

func NewGitScanner() *GitScanner {
	return &GitScanner{
		gitRepos: make([]string, 0),
	}
}

func (gs *GitScanner) ScanDirectory(rootPath string) ([]string, error) {
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == ".git" {
			gs.gitRepos = append(gs.gitRepos, filepath.Dir(path))
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return gs.gitRepos, nil
}