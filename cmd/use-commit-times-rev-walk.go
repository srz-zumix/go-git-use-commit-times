/*
Copyright © 2020 srz_zumix <https://github.com/srz-zumix>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/schollz/progressbar/v3"
)

func touch_files(workdir string, filemap FileIdMap, mtime time.Time) error {
	for path, _ := range filemap {
		fullpath := filepath.Join(workdir, path)
		if stat, err := os.Stat(fullpath); !os.IsNotExist(err) {
			if !stat.ModTime().Equal(mtime) {
				err := os.Chtimes(fullpath, mtime, mtime)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type ChtimeCallback = func(path string, mtime time.Time) error

func get_file_entries(commit *object.Commit, filemap FileIdMap, cb ChtimeCallback) (int, error) {
	count := 0
	tree, err := commit.Tree()
	if err != nil {
		return count, err
	}

	mtime := commit.Committer.When.UTC()

	// Handle initial commit (no parents)
	if commit.NumParents() == 0 {
		err = tree.Files().ForEach(func(f *object.File) error {
			if _, ok := filemap[f.Name]; ok {
				cb(f.Name, mtime)
				count++
				delete(filemap, f.Name)
			}
			return nil
		})
		return count, err
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
			if _, ok := filemap[path]; ok {
				cb(path, mtime)
				count++
				delete(filemap, path)
			}
		}
	}
	return count, nil
}

func get_file_entries_bybuf(commit *object.Commit, filemap FileIdMap, cb ChtimeCallback) (int, error) {
	return get_file_entries(commit, filemap, cb)
}

func get_file_entries_dummy(commit *object.Commit, filemap FileIdMap, cb ChtimeCallback) (int, error) {
	_ = commit
	_ = filemap
	_ = cb
	return 0, nil
}

func use_commit_times_rev_walk(repo *git.Repository, filemap FileIdMap, verbose bool, isShowProgress bool) error {
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

	onchtimes := func(path string, mtime time.Time) error {
		fullpath := filepath.Join(workdir, path)
		err := os.Chtimes(fullpath, mtime, mtime)
		if err != nil {
			return err
		}
		return nil
	}

	var m sync.Mutex

	// Get commit history in chronological order
	commitIter, err := repo.Log(&git.LogOptions{
		From: ref.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return err
	}

	err = commitIter.ForEach(func(commit *object.Commit) error {
		m.Lock()
		defer m.Unlock()

		if len(filemap) == 0 {
			return fmt.Errorf("done") // Error to break out of ForEach
		}

		count, err := get_file_entries_bybuf(commit, filemap, onchtimes)
		if err != nil {
			return err
		}
		if count > 0 && bar != nil {
			bar.Add(count)
		}
		return nil
	})

	// Treat "done" error as normal completion
	if err != nil && err.Error() != "done" {
		return err
	}

	if len(filemap) != 0 {
		fmt.Println("Warning: The final commit log for the file was not found.")
		if verbose {
			for k, _ := range filemap {
				fmt.Println(k)
			}
		}
		// err = touch_files(repo, filemap_to_entries(filemap), lastTime)
		// if err != nil {
		// 	return err
		// }
	}
	return nil
}
