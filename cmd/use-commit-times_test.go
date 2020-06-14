// +build !depth1

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
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/srz-zumix/git-use-commit-times/xgit"
)

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

	testfile := filepath.Join("..", path)
	now := time.Now()
	os.Chtimes(testfile, now, now)

	err = use_commit_times_rev_walk(repo, filemap, false)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	fileinfo, err := os.Stat(testfile)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	timeformat := "2006-01-02 15:04:05"
	expect, _ := time.Parse(timeformat, "2020-06-12 06:03:11")
	expect = expect.UTC()
	actual := fileinfo.ModTime().UTC()
	if expect != actual {
		t.Fatalf("failed modtime %s vs %s", expect.Format(timeformat), actual.Format(timeformat))
	}
}
