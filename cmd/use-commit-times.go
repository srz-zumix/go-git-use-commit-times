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

func get_file_entries(commit *git.Commit, diffOpts git.DiffOptions) ([]string, error) {
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}
	parent := commit.Parent(0)
	if parent == nil {
		return get_entries(tree)
	}
	parent_tree, err := parent.Tree()
	if err != nil {
		return nil, err
	}
	diff, err := commit.Owner().DiffTreeToTree(parent_tree, tree, &diffOpts)
	if err != nil {
		return nil, err
	}

	entries := []string{}
	diff.ForEach(func(file git.DiffDelta, _ float64) (git.DiffForEachHunkCallback, error) {
		entries = append(entries, file.NewFile.Path)
		return nil, nil
	}, git.DiffDetailFiles)

	return entries, nil
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

func use_commit_times(repo *git.Repository, files []string) error {
	filemap := make(map[string]struct{})
	for _, v := range files {
		filemap[v] = struct{}{}
	}
	rv, err := repo.Walk()
	if err != nil {
		return err
	}

	rv.Sorting(git.SortTime)
	err = rv.PushHead()
	if err != nil {
		return err
	}

	diffOpts, err := git.DefaultDiffOptions()
	if err != nil {
		return err
	}

	var oid git.Oid
	var lastTime time.Time
	for {
		err = rv.Next(&oid)
		if err != nil {
			break
		}
		obj, err := repo.Lookup(&oid)
		if err != nil {
			return err
		}
		if obj.Type() == git.ObjectCommit {
			commit, err := obj.AsCommit()
			if err != nil {
				return err
			}
			lastTime = commit.Committer().When.UTC()
			git_print_commit(commit)
			fmt.Println(lastTime.Format("2006-01-02 15:04:05 MST"))

			entries, err := get_file_entries(commit, diffOpts)
			if err != nil {
				return err
			}

			// fmt.Println(entries)
			filemap, entries = update_files(filemap, entries)
			if len(entries) > 0 {
				fmt.Println(entries)
				err = touch_files(repo, entries, lastTime)
				if err != nil {
					return err
				}
			}
			if len(filemap) == 0 {
				return nil
			}
		}
	}

	if len(filemap) != 0 {
		fmt.Println("Warning: The final commit log for the file was not found.")
		err = touch_files(repo, filemap_to_entries(filemap), lastTime)
		if err != nil {
			return err
		}
	}
	return nil
}
