//go:build !depth1

package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
)

func TestMTimeGitLog(t *testing.T) {
	repo, err := git.PlainOpen("../")
	if err != nil {
		t.Fatalf("failed to open repository: %#v", err)
	}
	path := "tests/testfile"
	files := []string{path}
	filemap, err := get_fileidmap(repo, files)
	if err != nil {
		t.Fatalf("failed to get fileidmap for %s: %#v", path, err)
	}

	testfile := filepath.Join("..", path)
	now := time.Now()
	os.Chtimes(testfile, now, now)

	err = use_commit_times_log_walk(repo, filemap, nil, nil, false)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	fileinfo, err := os.Stat(testfile)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	timeformat := "2006-01-02 15:04:05"
	outformat := "2006-01-02 15:04:05 MST"
	expect, _ := time.Parse(timeformat, "2020-06-12 06:03:11")
	expect = expect.UTC()
	actual := fileinfo.ModTime().UTC()
	if expect != actual {
		t.Fatalf("failed modtime %s vs %s", expect.Format(outformat), actual.Format(outformat))
	}
}
