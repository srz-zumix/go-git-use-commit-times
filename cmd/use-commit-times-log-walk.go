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

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/schollz/progressbar/v3"
)

func use_commit_times_log_walk(repo *git.Repository, filemap FileIdMap, since string, verbose bool, isShowProgress bool) error {
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

	var lastTime time.Time
	hasLastTime := false

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

	// コミット履歴を時系列順に取得
	commitIter, err := repo.Log(&git.LogOptions{
		From: ref.Hash(),
		Order: git.LogOrderCommitterTime,
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
		hasLastTime = true

		// since オプションのチェック
		if since != "" {
			// 簡易的な時刻パース - より厳密な実装が必要な場合は time.Parse を使用
			// ここでは簡略化のため省略
		}

		tree, err := commit.Tree()
		if err != nil {
			return err
		}

		// 親コミットがない場合（初回コミット）
		if commit.NumParents() == 0 {
			err = tree.Files().ForEach(func(f *object.File) error {
				chtimes(f.Name, mtime)
				return nil
			})
			return err
		}

		// 各親コミットとの差分を確認
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

	// "done"エラーは正常終了として扱う
	if err != nil && err.Error() != "done" {
		return err
	}

	if len(filemap) != 0 {
		fmt.Println("Warning: The final commit log for the file was not found.")
		if verbose {
			for k := range filemap {
				fmt.Println(k)
			}
		}
		if hasLastTime {
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
	}
	return nil
}
