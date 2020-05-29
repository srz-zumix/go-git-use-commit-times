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
	"path/filepath"

	git "github.com/srz-zumix/git-use-commit-times/xgit"
)

func ls_files(repo *git.Repository) ([]string, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}
	obj, err := ref.Peel(git.ObjectTree)
	if err != nil {
		return nil, err
	}

	tree, err := obj.AsTree()
	if err != nil {
		return nil, err
	}
	return get_entries(tree)
}

func get_entries(tree *git.Tree) ([]string, error) {
	files := []string{}
	callback := func(e string, te *git.TreeEntry) int {
		// fmt.Println("e: ", e, ", te: ", te.Name)
		files = append(files, filepath.Join(e, te.Name))
		return 0
	}
	tree.Walk(callback)
	return files, nil
}
