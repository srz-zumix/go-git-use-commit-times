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

	"github.com/schollz/progressbar/v3"
	git "github.com/srz-zumix/git-use-commit-times/xgit"
)

func update_time_byid(repo *git.Repository, path string, id *git.Oid) error {
	obj, err := repo.Lookup(id)
	if err != nil {
		return err
	}

	var commit *git.Commit = nil
	switch obj.Type() {
	case git.ObjectCommit:
		commit, err = obj.AsCommit()
		if err != nil {
			return err
		}
		defer commit.Free()
		// default:
		// 	fmt.Println(obj.Type())
	}

	if commit == nil {
		return fmt.Errorf("commit not found")
	}
	lastTime := commit.Committer().When.UTC()
	os.Chtimes(path, lastTime, lastTime)
	return nil
}

func use_commit_times_tree_walk(repo *git.Repository, filemap FileIdMap, isShowProgress bool) error {
	total := int64(len(filemap))
	current := 0
	var bar *progressbar.ProgressBar = nil
	if isShowProgress {
		bar = progressbar.Default(total)
	}

	workdir := repo.Workdir()

	updater := func(path string, id *git.Oid) error {
		err := update_time_byid(repo, filepath.Join(workdir, path), id)
		if err != nil {
			return err
		}
		current++
		if bar != nil {
			bar.Add(1)
		}
		return nil
	}

	for k, v := range filemap {
		err := updater(k, v)
		if err != nil {
			return err
		}
		// go updater(k, v)
	}

	if bar != nil {
		bar.Finish()
	}
	return nil
}
