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
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
	git "github.com/srz-zumix/git-use-commit-times/xgit"
)

func filemap_to_entries(filemap map[string]struct{}) []string {
	files := make([]string, len(filemap))
	for k, _ := range filemap {
		files = append(files, k)
	}
	return files
}

func touch_files(workdir string, entries []string, mtime time.Time) error {
	for _, path := range entries {
		err := os.Chtimes(filepath.Join(workdir, path), mtime, mtime)
		if err != nil {
			return err
		}
	}
	return nil
}

type ChtimeCallback = func(path string, mtime time.Time) error

func get_file_entries(commit *git.Commit, filemap FileIdMap, cb ChtimeCallback) (int, error) {
	tree, err := commit.Tree()
	count := 0
	if err != nil {
		return count, err
	}
	defer tree.Free()

	for i := uint(0); i < commit.ParentCount(); i++ {
		parent := commit.Parent(i)
		parent_tree, err := parent.Tree()
		if err != nil {
			return count, err
		}
		diff, err := commit.Owner().DiffTreeToTree(parent_tree, tree, nil)
		if err != nil {
			return count, err
		}
		mtime := commit.Committer().When.UTC()

		diff.ForEach(func(file git.DiffDelta, _ float64) (git.DiffForEachHunkCallback, error) {
			switch file.Status {
			case git.DeltaAdded:
				fallthrough
			case git.DeltaModified:
				fallthrough
			case git.DeltaRenamed:
				fallthrough
			case git.DeltaCopied:
				fallthrough
			case git.DeltaTypeChange:
				path := file.NewFile.Path
				if _, ok := filemap[path]; ok {
					cb(path, mtime)
					count++
					delete(filemap, path)
				}
			}
			return nil, nil
		}, git.DiffDetailFiles)

		diff.Free()
		parent.Free()
		parent_tree.Free()
	}
	return count, nil
}

func get_file_entries_bybuf(commit *git.Commit, filemap FileIdMap, cb ChtimeCallback) (int, error) {
	count := 0
	tree, err := commit.Tree()
	if err != nil {
		return count, err
	}
	defer tree.Free()
	mtime := commit.Committer().When.UTC()

	for i := uint(0); i < commit.ParentCount(); i++ {
		parent := commit.Parent(i)
		parent_tree, err := parent.Tree()
		if err != nil {
			return count, err
		}
		diff, err := commit.Owner().DiffTreeToTree(parent_tree, tree, nil)
		if err != nil {
			return count, err
		}

		buf, err := git.DiffToBuf(diff, git.DiffFormatNameOnly)
		if err != nil {
			return count, err
		}
		for _, path := range strings.Split(string(buf), "\n") {
			if _, ok := filemap[path]; ok {
				cb(path, mtime)
				count++
				delete(filemap, path)
			}
		}
	}
	return count, nil
}

func get_file_entries_dummy(commit *git.Commit, filemap FileIdMap, cb ChtimeCallback) (int, error) {
	_ = commit
	_ = filemap
	_ = cb
	return 0, nil
}

func use_commit_times_rev_walk(repo *git.Repository, filemap FileIdMap, isShowProgress bool) error {
	// defer profile.Start(profile.ProfilePath(".")).Stop()

	total := int64(len(filemap))
	current := 0
	var bar *progressbar.ProgressBar = nil
	if isShowProgress {
		bar = progressbar.Default(total)
	}

	workdir := repo.Workdir()
	rv, err := repo.Walk()
	if err != nil {
		return err
	}
	defer rv.Free()

	// rv.Sorting(git.SortTime)
	rv.Sorting(git.SortNone)
	err = rv.PushHead()
	if err != nil {
		return err
	}

	onchtimes := func(path string, mtime time.Time) error {
		err := os.Chtimes(filepath.Join(workdir, path), mtime, mtime)
		if err != nil {
			return err
		}
		return nil
	}

	var m sync.Mutex

	onvisit := func(commit *git.Commit) bool {
		m.Lock()
		if len(filemap) == 0 {
			return false
		}

		go func(commit *git.Commit) {
			defer m.Unlock()
			count, err := get_file_entries_bybuf(commit, filemap, onchtimes)
			if err != nil {
				return
			}
			if count > 0 {
				if bar != nil {
					bar.Add(count)
				}
				current += count
			}
		}(commit)
		return true
	}

	err = rv.Iterate(onvisit)
	if err != nil {
		return err
	}
	if len(filemap) != 0 {
		fmt.Println("Warning: The final commit log for the file was not found.")
		// err = touch_files(repo, filemap_to_entries(filemap), lastTime)
		// if err != nil {
		// 	return err
		// }
	}
	if bar != nil {
		bar.Finish()
	}
	return nil
}
