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
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
	git "github.com/srz-zumix/git-use-commit-times/xgit"
)

func use_commit_times_log_walk(repo *git.Repository, filemap FileIdMap, isShowProgress bool) error {
	total := int64(len(filemap))
	var bar *progressbar.ProgressBar = nil
	if isShowProgress {
		bar = progressbar.Default(total)
		defer bar.Finish()
	}

	workdir := repo.Workdir()

	rcommitter := regexp.MustCompile(`committer .*? (\d+) ([\-\+]\d+)$`)
	rcommit := regexp.MustCompile(`commit [a-f0-9]{40}( \(from [a-f0-9]{40}\))?$`)
	var lastTime time.Time
	hasLastTime := false

	// var wg sync.WaitGroup
	var mtx sync.Mutex
	chtimes := func(files []string, lastTime time.Time) {
		// defer mtx.Unlock()
		count := 0
		for _, path := range files {
			if path != "" {
				if _, ok := filemap[path]; ok {
					// wg.Add(1)
					// go func(path string, lastTime time.Time) {
					// 	os.Chtimes(path, lastTime, lastTime)
					// 	wg.Done()
					// }(filepath.Join(workdir, path), lastTime)
					os.Chtimes(filepath.Join(workdir, path), lastTime, lastTime)
					count++
					delete(filemap, path)
				}
				// if _, err := os.Stat(filepath.Join(workdir, path)); err != nil {
				// 	fmt.Println(path)
				// }
			}
		}
		if bar != nil {
			bar.Add(count)
		}
	}
	onvisit := func(line string) bool {
		mtx.Lock()
		defer mtx.Unlock()
		// if strings.Index(line, "tree") == 0 {
		// 	fmt.Println(line)
		// }
		if strings.Index(line, "committer") == 0 {
			fmt.Println(line)
			m := rcommitter.FindStringSubmatch(line)
			unix, _ := strconv.ParseInt(m[1], 10, 64)
			lastTimeUTC := time.Unix(unix, 0).UTC()
			lastTime, _ = time.Parse("2006-01-02 15:04:05 -0700", lastTimeUTC.Format("2006-01-02 15:04:05")+" "+m[2])
			hasLastTime = true
		} else if strings.Index(line, "\x00") >= 0 {
			// fmt.Println(line)
			line = rcommit.ReplaceAllString(line, "")
			files := strings.Split(line, "\x00")
			// for _, f := range files {
			// 	fmt.Println(f)
			// }
			chtimes(files, lastTime)
			if len(filemap) == 0 {
				return true
			}
		}
		return false
	}

	args := []string{"--no-pager", "-c", "diff.renames=false", "log", "-m", "-r", "--name-only", "--no-color", "--pretty=raw", "-z"}
	cmd := exec.Command("git", args...)
	cmd.Dir = workdir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()

	streamReader := func(reader *bufio.Reader, outputChan chan string, doneChan chan bool) {
		defer close(outputChan)
		defer close(doneChan)
		for true {
			// buf, isPrefix, err := reader.ReadLine()
			// bb := bytes.NewBuffer(buf)
			// if isPrefix {
			// 	for {
			// 		b, cont, err := reader.ReadLine()
			// 		_, werr := bb.Write(b)
			// 		if werr != nil {
			// 			log.Fatal(err)
			// 		}

			// 		if err != nil {
			// 			break
			// 		}
			// 		if !cont {
			// 			break
			// 		}
			// 	}
			// }
			// line := string(bb.Bytes())
			// if len(line) > 0 {
			// 	outputChan <- line
			// }

			line, err := reader.ReadString('\n')
			line = strings.TrimRight(line, "\n")
			if len(line) > 0 {
				outputChan <- line
			}

			if err != nil {
				if err != io.EOF {
					fmt.Println(err)
				}
				break
			}
		}
		// for scanner.Scan() {
		// 	outputChan <- scanner.Text()
		// }
		doneChan <- true
	}
	// stdoutScanner := bufio.NewScanner(stdout)
	stdoutReader := bufio.NewReader(stdout)
	stdoutOutputChan := make(chan string)
	stdoutDoneChan := make(chan bool)

	go streamReader(stdoutReader, stdoutOutputChan, stdoutDoneChan)

	stillGoing := true
	for stillGoing {
		select {
		case line := <-stdoutOutputChan:
			if onvisit(line) {
				stillGoing = false
				cmd.Process.Kill()
			}
		case <-stdoutDoneChan:
			stillGoing = false
			cmd.Process.Kill()
			// fmt.Println(stdoutScanner.Err())
		}
	}

	cmd.Wait()
	// wg.Wait()

	if len(filemap) != 0 {
		fmt.Println("Warning: The final commit log for the file was not found.")
		for k, _ := range filemap {
			fmt.Println(k)
		}
		if hasLastTime {
			err = touch_files(workdir, filemap, lastTime)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
