package repo

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog/log"
	"time"
)

func CommitToRepo(gitRepoDir string, commitMsg string, commitUser string, commitEmail string) (err error) {
	// References:
	// https://github.com/go-git/go-git/tree/master/_examples
	// https://github.com/ad-m/github-push-action/blob/master/start.js

	log.Info().Msgf("Opening Git repository at %v", gitRepoDir)
	repo, err := git.PlainOpen(gitRepoDir)
	if err != nil {
		return
	}

	w, err := repo.Worktree()
	if err != nil {
		return
	}

	log.Info().Msg("Checking status of working tree")
	status, err := w.Status()
	if err != nil {
		return
	}

	if status.IsClean() {
		log.Info().Msg("🏆 No changes to commit")
	} else {
		log.Info().Msg("Adding all files for Git tracking")
		err = w.AddWithOptions(&git.AddOptions{All: true})
		if err != nil {
			return
		}

		status, err = w.Status()
		if err != nil {
			return
		}
		log.Info().Msgf("Status of working tree:\n%v", status)

		log.Info().Msg("Trying to commit changes")
		var commit plumbing.Hash
		commit, err = w.Commit(commitMsg, &git.CommitOptions{
			All: true,
			Author: &object.Signature{
				Name:  commitUser,
				Email: commitEmail,
				When:  time.Now(),
			},
		})
		if err != nil {
			return
		}

		var obj *object.Commit
		obj, err = repo.CommitObject(commit)
		if err != nil {
			return
		}

		log.Info().Msgf("Commit object:\n%v", obj)
		log.Info().Msg("🏆 Changes committed")
	}
	return
}
