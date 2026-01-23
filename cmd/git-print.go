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

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func git_print_tree(tree *object.Tree) error {
	fmt.Printf("obj:  %v\n", tree)
	fmt.Printf("Hash: %s\n", tree.Hash)
	fmt.Printf("EntryCount: %d\n", len(tree.Entries))
	for _, entry := range tree.Entries {
		fmt.Println("  ", entry.Name, entry.Hash)
	}
	return nil
}

func git_print_tree_from_commit(commit *object.Commit) error {
	tree, err := commit.Tree()
	if err != nil {
		return err
	}
	return git_print_tree(tree)
}

func git_print_blob(repo *git.Repository, hash plumbing.Hash) error {
	blob, err := repo.BlobObject(hash)
	if err != nil {
		return err
	}
	fmt.Printf("obj:  %v\n", blob)
	fmt.Printf("Hash: %s\n", blob.Hash)
	fmt.Printf("Size: %d\n", blob.Size)
	return nil
}

func git_print_commit(commit *object.Commit) error {
	fmt.Printf("obj:  %v\n", commit)
	fmt.Printf("Hash: %s\n", commit.Hash)
	author := commit.Author
	fmt.Printf("    Author:\n        Name:  %s\n        Email: %s\n        Date:  %s\n", author.Name, author.Email, author.When)
	committer := commit.Committer
	fmt.Printf("    Committer:\n        Name:  %s\n        Email: %s\n        Date:  %s\n", committer.Name, committer.Email, committer.When)
	fmt.Printf("    ParentCount: %d\n", commit.NumParents())
	fmt.Printf("    TreeHash:    %s\n", commit.TreeHash)
	fmt.Printf("    Message:\n\n        %s\n\n", strings.Replace(commit.Message, "\n", "\n        ", -1))
	fmt.Printf("--------------\n")
	return nil
}

func git_print_odb(repo *git.Repository) error {
	ref, err := repo.Head()
	if err != nil {
		return err
	}

	commitIter, err := repo.Log(&git.LogOptions{
		From: ref.Hash(),
	})
	if err != nil {
		return err
	}

	err = commitIter.ForEach(func(commit *object.Commit) error {
		fmt.Printf("==================================\n")
		git_print_commit(commit)
		git_print_tree_from_commit(commit)
		return nil
	})

	return err
}
