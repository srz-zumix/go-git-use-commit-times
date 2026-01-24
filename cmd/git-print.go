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
