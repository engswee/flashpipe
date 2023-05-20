package repo

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"log"
	"time"
)

func CommitToRepo(gitSrcDir string, commitMsg string) {
	fmt.Println("[INFO] Opening Git repository at", gitSrcDir)
	repo, err := git.PlainOpen(gitSrcDir)
	CheckIfError(err)

	w, err := repo.Worktree()
	CheckIfError(err)

	fmt.Println("[INFO] Checking status of working tree")
	status, err := w.Status()
	CheckIfError(err)

	if status.IsClean() {
		fmt.Println("[INFO] 🏆 No changes to commit")
	} else {
		fmt.Println("[INFO] Adding all files for Git tracking")
		err = w.AddWithOptions(&git.AddOptions{All: true})
		CheckIfError(err)

		status, err = w.Status()
		CheckIfError(err)
		fmt.Println(status)

		fmt.Println("[INFO] Trying to commit changes")
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
		fmt.Println("[INFO] 🏆 Changes committed")
	}
}

// Info should be used to describe the example commands that are about to run.
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

func CheckIfError(err error) {
	if err != nil {
		//fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
		log.SetFlags(0)
		log.Fatalln("[ERROR] 🛑", err)
	}
}
