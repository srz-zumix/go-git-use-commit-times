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
	"strings"

	git "github.com/srz-zumix/git-use-commit-times/xgit"
)

func git_print_tree(tree *git.Tree) error {
	fmt.Printf("obj:  %s\n", tree)
	fmt.Printf("Type: %s\n", tree.Type())
	fmt.Printf("Id:   %s\n", tree.Id())
	fmt.Printf("EntryCount: %s\n", tree.EntryCount())
	callback := func(e string, te *git.TreeEntry) int {
		fmt.Println("  ", e, te.Name)
		return 0
	}
	tree.Walk(callback)
	return nil
}

func git_print_tree_from_commit(commit *git.Commit) error {
	tree, err := commit.Tree()
	if err != nil {
		return err
	}
	return git_print_tree(tree)
}

func git_print_tree_from_object(obj *git.Object) error {
	switch obj.Type() {
	case git.ObjectTree:
		tree, err := obj.AsTree()
		if err != nil {
			return err
		}
		return git_print_tree(tree)
	case git.ObjectCommit:
		commit, err := obj.AsCommit()
		if err != nil {
			return err
		}
		return git_print_tree_from_commit(commit)
	}
	return nil
}

func git_print_blob(blob *git.Blob) error {
	fmt.Printf("obj:  %s\n", blob)
	fmt.Printf("Type: %s\n", blob.Type())
	fmt.Printf("Id:   %s\n", blob.Id())
	fmt.Printf("Size: %s\n", blob.Size())
	return nil
}

func git_print_blob_from_object(obj *git.Object) error {
	blob, err := obj.AsBlob()
	if err != nil {
		return err
	}
	return git_print_blob(blob)
}

func git_print_commit(commit *git.Commit) error {
	fmt.Printf("obj:  %s\n", commit)
	fmt.Printf("Type: %s\n", commit.Type())
	fmt.Printf("Id:   %s\n", commit.Id())
	author := commit.Author()
	fmt.Printf("    Author:\n        Name:  %s\n        Email: %s\n        Date:  %s\n", author.Name, author.Email, author.When)
	committer := commit.Committer()
	fmt.Printf("    Committer:\n        Name:  %s\n        Email: %s\n        Date:  %s\n", committer.Name, committer.Email, committer.When)
	fmt.Printf("    ParentCount: %s\n", int(commit.ParentCount()))
	fmt.Printf("    TreeId:      %s\n", commit.TreeId())
	fmt.Printf("    Message:\n\n        %s\n\n", strings.Replace(commit.Message(), "\n", "\n        ", -1))
	fmt.Printf("--------------\n")
	return nil
}

func git_print_commit_from_object(obj *git.Object) error {
	commit, err := obj.AsCommit()
	if err != nil {
		return err
	}
	return git_print_commit(commit)
}

func git_print_odb(repo *git.Repository) error {
	odb, err := repo.Odb()
	if err != nil {
		return err
	}
	err = odb.ForEach(func(oid *git.Oid) error {
		obj, err := repo.Lookup(oid)
		if err != nil {
			return err
		}
		switch obj.Type() {
		default:
			break
		case git.ObjectBlob:
			fmt.Printf("==================================\n")
			git_print_blob_from_object(obj)
		case git.ObjectCommit:
			fmt.Printf("==================================\n")
			git_print_commit_from_object(obj)
			git_print_tree_from_object(obj)
		case git.ObjectTree:
			fmt.Printf("==================================\n")
			git_print_tree_from_object(obj)
		}
		return nil
	})
	return err
}
