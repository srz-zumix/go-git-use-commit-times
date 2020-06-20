// +build depth1

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
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	git "github.com/srz-zumix/git-use-commit-times/xgit"
)

func captureStdout(f func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	stdout := os.Stdout
	os.Stdout = w

	f()

	os.Stdout = stdout
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}

func TestMTime(t *testing.T) {
	repo, err := git.OpenRepository("../")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	path := "tests/testfile"
	files := []string{path}
	filemap, err := get_fileidmap(repo, files)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	err = use_commit_times_rev_walk(repo, filemap, false, false)
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
	commit, err := repo.LookupCommit(ref.Target())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	timeformat := "2006-01-02 15:04:05 MST"
	expect := commit.Committer().When.UTC()
	actual := fileinfo.ModTime().UTC()
	if expect != actual {
		t.Fatalf("failed modtime %s vs %s", expect.Format(timeformat), actual.Format(timeformat))
	}
}
