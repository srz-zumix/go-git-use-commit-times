package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/schollz/progressbar/v3"
)

func use_commit_times_log_walk(repo *git.Repository, filemap FileIdMap, since *time.Time, until *time.Time, isShowProgress bool) error {
	Logger.Info("Starting commit time update (log walk)", "files", len(filemap), "since", since, "until", until)
	total := int64(len(filemap))
	var bar *progressbar.ProgressBar = nil
	if isShowProgress {
		bar = progressbar.Default(total)
		defer bar.Finish()
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}
	workdir := worktree.Filesystem.Root()

	ref, err := repo.Head()
	if err != nil {
		return err
	}

	// Get lastTime from HEAD commit
	headCommit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return err
	}
	lastTime := headCommit.Committer.When

	chtimes := func(path string, mtime time.Time) {
		if _, ok := filemap[path]; ok {
			fullpath := filepath.Join(workdir, path)
			stat, err := os.Stat(fullpath)
			if err == nil {
				if !stat.ModTime().Equal(mtime) {
					os.Chtimes(fullpath, mtime, mtime)
				}
			}
			delete(filemap, path)
			if bar != nil {
				bar.Add(1)
			}
		}
	}

	// Get commit history in chronological order
	commitIter, err := repo.Log(&git.LogOptions{
		From: ref.Hash(),
		Order: git.LogOrderCommitterTime,
		Since: since,
		Until: until,
	})
	if err != nil {
		return err
	}

	err = commitIter.ForEach(func(commit *object.Commit) error {
		if len(filemap) == 0 {
			Logger.Debug("All files processed, stopping iteration")
			return fmt.Errorf("done")
		}

		Logger.Debug("Processing commit", "hash", commit.Hash, "remaining", len(filemap))
		mtime := commit.Committer.When
		lastTime = mtime

		tree, err := commit.Tree()
		if err != nil {
			return err
		}

		// Handle initial commit (no parents)
		if commit.NumParents() == 0 {
			err = tree.Files().ForEach(func(f *object.File) error {
				chtimes(f.Name, mtime)
				return nil
			})
			return err
		}

		// Check diff with each parent commit
		for i := 0; i < commit.NumParents(); i++ {
			parent, err := commit.Parent(i)
			if err != nil {
				continue
			}

			parentTree, err := parent.Tree()
			if err != nil {
				continue
			}

			changes, err := parentTree.Diff(tree)
			if err != nil {
				continue
			}

			for _, change := range changes {
				path := change.To.Name
				if path == "" {
					path = change.From.Name
				}
				chtimes(path, mtime)
			}
		}

		return nil
	})

	// Treat "done" error as normal completion
	if err != nil && err.Error() != "done" {
		Logger.Warn("Error during commit iteration", "error", err)
	}

	if len(filemap) != 0 {
		Logger.Warn("Some files not found in commit history", "count", len(filemap))
		for k := range filemap {
			Logger.Warn("File not found", "path", k)
		}
		// Use lastTime (from HEAD) for remaining files
		for path := range filemap {
			fullpath := filepath.Join(workdir, path)
			stat, err := os.Stat(fullpath)
			if err == nil {
				if !stat.ModTime().Equal(lastTime) {
					os.Chtimes(fullpath, lastTime, lastTime)
				}
			}
		}
	}
	return nil
}
