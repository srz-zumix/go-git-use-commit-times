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
	"time"

	"github.com/schollz/progressbar/v3"
	git "github.com/srz-zumix/git-use-commit-times/xgit"
)

func update_files(filemap map[string]struct{}, entries []string) (map[string]struct{}, []string) {
	matches := []string{}
	for _, e := range entries {
		if _, ok := filemap[e]; ok {
			matches = append(matches, e)
			delete(filemap, e)
		}
	}
	return filemap, matches
}

func filemap_to_entries(filemap map[string]struct{}) []string {
	files := []string{}
	for k, _ := range filemap {
		files = append(files, k)
	}
	return files
}

func get_file_entries(commit *git.Commit, diffOpts git.DiffOptions, filemap map[string]struct{}) (map[string]struct{}, []string, error) {
	tree, err := commit.Tree()
	if err != nil {
		return filemap, nil, err
	}
	defer tree.Free()

	entries := []string{}
	for i := uint(0); i < commit.ParentCount(); i++ {
		parent := commit.Parent(i)
		parent_tree, err := parent.Tree()
		if err != nil {
			return filemap, nil, err
		}
		diff, err := commit.Owner().DiffTreeToTree(parent_tree, tree, &diffOpts)
		if err != nil {
			return filemap, nil, err
		}

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
					entries = append(entries, path)
					delete(filemap, path)
				}
			}
			return nil, nil
		}, git.DiffDetailFiles)

		diff.Free()
		parent.Free()
		parent_tree.Free()
	}
	return filemap, entries, nil
}

func touch_files(repo *git.Repository, entries []string, mtime time.Time) error {
	workdir := repo.Workdir()
	for _, path := range entries {
		err := os.Chtimes(filepath.Join(workdir, path), mtime, mtime)
		if err != nil {
			return err
		}
	}
	return nil
}

func use_commit_times(repo *git.Repository, files []string, isShowProgress bool) error {
	total := int64(len(files))
	current := 0
	var bar *progressbar.ProgressBar = nil
	if isShowProgress {
		bar = progressbar.Default(total)
	}
	filemap := make(map[string]struct{})
	for _, v := range files {
		filemap[v] = struct{}{}
	}
	rv, err := repo.Walk()
	if err != nil {
		return err
	}

	// rv.Sorting(git.SortTime)
	err = rv.PushHead()
	if err != nil {
		return err
	}

	diffOpts, err := git.DefaultDiffOptions()
	if err != nil {
		return err
	}
	diffOpts.IgnoreSubmodules = git.SubmoduleIgnoreAll

	var lastTime time.Time

	oid := new(git.Oid)
	for {
		err = rv.Next(oid)
		if err != nil {
			if git.IsErrorCode(err, git.ErrIterOver) {
				break
			}
			return err
		}

		commit, err := repo.LookupCommit(oid)
		if err != nil {
			return err
		}

		lastTime = commit.Committer().When.UTC()
		// git_print_commit(commit)
		// fmt.Println(lastTime.Format("2006-01-02 15:04:05 MST"))

		filemap, entries, err := get_file_entries(commit, diffOpts, filemap)

		// fmt.Printf("%d/%d\n", current, total)
		// fmt.Printf("tree %s\n", commit.TreeId())

		count := len(entries)
		if count > 0 {
			// fmt.Println(count)
			// fmt.Println(entries)
			// fmt.Println(strings.Join(entries, "\n"))
			err = touch_files(repo, entries, lastTime)
			if err != nil {
				return err
			}
			// go touch_files(repo, entries, lastTime)
			// _ = lastTime
			if bar != nil {
				bar.Add(count)
			}
			current += count
		}
		if len(filemap) == 0 {
			break
		}
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
