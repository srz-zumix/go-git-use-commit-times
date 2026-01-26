//go:build usegogit

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func chtimes(workdir string, path string, mtime time.Time) error {
	fullpath := filepath.Join(workdir, path)
	stat, err := os.Stat(fullpath)
	if err == nil {
		if !stat.ModTime().Equal(mtime) {
			return os.Chtimes(fullpath, mtime, mtime)
		}
	}
	return nil
}

func use_commit_times_walk(repo *git.Repository, filemap FileIdMap, since *time.Time, until *time.Time) error {
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

	on_chtimes := func(path string, mtime time.Time) error {
		if _, ok := filemap[path]; ok {
			err := chtimes(workdir, path, mtime)
			if err != nil {
				return err
			}
			delete(filemap, path)
		}
			return nil
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
			return fmt.Errorf("done")
		}

		mtime := commit.Committer.When
		lastTime = mtime

		tree, err := commit.Tree()
		if err != nil {
			return err
		}

		// Handle initial commit (no parents)
		if commit.NumParents() == 0 {
			err = tree.Files().ForEach(func(f *object.File) error {
				return on_chtimes(f.Name, mtime)
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

			changes, err := object.DiffTreeWithOptions(context.Background(), parentTree, tree, &object.DiffTreeOptions{
				DetectRenames: false,
			})
			if err != nil {
				continue
			}

			for _, change := range changes {
				// Process only additions and modifications (not deletions)
				if change.To.Name != "" {
					err := on_chtimes(change.To.Name, mtime)
					if err != nil {
						return err
					}
				}
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
			err := on_chtimes(path, lastTime)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
