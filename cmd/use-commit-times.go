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

func use_commit_times(repo *git.Repository, files []string) error {
	filemap := make(map[string]struct{})
	for _, v := range files {
		filemap[v] = struct{}{}
	}
	odb, err := repo.Odb()
	if err != nil {
		return err
	}
	err = odb.ForEach(func(oid *git.Oid) error {
		obj, err := repo.Lookup(oid)
		if err != nil {
			return err
		}
		if obj.Type() == git.ObjectCommit {
			commit, err := obj.AsCommit()
			if err != nil {
				return err
			}
			tree, err := commit.Tree()
			if err != nil {
				return err
			}
			entries, err := get_entries(tree)
			if err != nil {
				return err
			}

			filemap, entries = update_files(filemap, entries)
			if len(entries) > 0 {
				// fmt.Println(entries)
				time := commit.Committer().When
				for _, path := range entries {
					os.Chtimes(path, time, time)
					// fileinfo, err := os.Stat(path)
					// if err != nil {
					// 	return err
					// }
					// fmt.Println(fileinfo.ModTime())
				}
			}
			if len(filemap) == 0 {
				return fmt.Errorf("break")
			}
		}
		return nil
	})
	return nil
}
