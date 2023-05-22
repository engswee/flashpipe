package repo

import (
	"fmt"
	"github.com/engswee/flashpipe/logger"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"time"
)

func CommitToRepo(gitSrcDir string, commitMsg string) {
	logger.Info("Opening Git repository at", gitSrcDir)
	repo, err := git.PlainOpen(gitSrcDir)
	CheckIfError(err)

	w, err := repo.Worktree()
	CheckIfError(err)

	logger.Info("Checking status of working tree")
	status, err := w.Status()
	CheckIfError(err)

	if status.IsClean() {
		logger.Info("🏆 No changes to commit")
	} else {
		logger.Info("Adding all files for Git tracking")
		err = w.AddWithOptions(&git.AddOptions{All: true})
		CheckIfError(err)

		status, err = w.Status()
		CheckIfError(err)
		fmt.Println(status)

		logger.Info("Trying to commit changes")
		commit, err := w.Commit(commitMsg, &git.CommitOptions{
			All: true,
			Author: &object.Signature{
				Name:  "github-actions[bot]", // TODO - switch to flag
				Email: "41898282+github-actions[bot]@users.noreply.github.com",
				When:  time.Now(),
			},
		})
		CheckIfError(err)

		obj, err := repo.CommitObject(commit)
		CheckIfError(err)

		fmt.Println(obj)
		logger.Info("🏆 Changes committed")
	}
}

func CheckIfError(err error) {
	if err != nil {
		logger.Error(err)
	}
}
