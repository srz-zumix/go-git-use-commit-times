//go:build depth1

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
)

func TestMTime(t *testing.T) {
	repo, err := git.PlainOpen("..")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %#v", err)
	}
	workdir := worktree.Filesystem.Root()

	path := "tests/testfile"
	files := []string{path}
	filemap, err := get_fileidmap(workdir, files)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	err = use_commit_times_walk(workdir, filemap, nil, nil)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	fileinfo, err := os.Stat(filepath.Join("..", path))
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	ref, err := repo.Head()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	timeformat := "2006-01-02 15:04:05 MST"
	expect := commit.Committer.When.UTC()
	actual := fileinfo.ModTime().UTC()
	if expect != actual {
		t.Fatalf("failed modtime %s vs %s", expect.Format(timeformat), actual.Format(timeformat))
	}
}
