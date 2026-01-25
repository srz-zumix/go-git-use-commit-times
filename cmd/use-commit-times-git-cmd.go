//go:build !usegogit

package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
)

// use_commit_times_git_cmd implements the same logic as the Perl script
// https://gist.github.com/srz-zumix/0a526e8f9182549cbdb6d880a4477ff0
// using git command line instead of go-git library
func use_commit_times_walk(repo *git.Repository, filemap FileIdMap, since *time.Time, until *time.Time) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}
	workdir := worktree.Filesystem.Root()

	// Pre-calculate full paths for all files
	filePaths := make(map[string]string, len(filemap))
	for file := range filemap {
		filePaths[file] = filepath.Join(workdir, file)
	}

	// Build git log command arguments
	args := []string{"-c", "diff.renames=false", "log", "-m", "-r", "--name-only", "--no-color", "--pretty=raw", "-z"}

	// Add time range if specified
	if since != nil {
		args = append(args, fmt.Sprintf("--since=%s", since.Format(time.RFC3339)))
	}
	if until != nil {
		args = append(args, fmt.Sprintf("--until=%s", until.Format(time.RFC3339)))
	}

	// Execute git log command with streaming output
	cmd := exec.Command("git", args...)
	cmd.Dir = workdir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start git log: %w", err)
	}

	// Use worker pool for concurrent file time updates
	type fileUpdate struct {
		path  string
		mtime time.Time
	}

	numWorkers := runtime.NumCPU()
	updateChan := make(chan fileUpdate, numWorkers*10)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for update := range updateChan {
				os.Chtimes(update.path, update.mtime, update.mtime)
			}
		}()
	}

	// Parse git log output with streaming
	var commitTime time.Time
	var lastCommitTime time.Time

	reader := bufio.NewReaderSize(stdout, 256*1024) // 256KB buffer
	lineBuffer := make([]byte, 0, 4096)

	for len(filemap) > 0 {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading git log output: %w", err)
		}

		// Trim newline
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}

		// Fast check for committer line (starts with 'c')
		if len(line) > 10 && line[0] == 'c' && bytes.HasPrefix(line, []byte("committer ")) {
			// Parse timestamp directly without string conversion
			if ts := parseCommitterTimestampFast(line); ts > 0 {
				commitTime = time.Unix(ts, 0)
				lastCommitTime = commitTime
			}
		} else if len(line) > 0 {
			// Check if line contains file names (has null bytes or ends with null)
			if bytes.Contains(line, []byte("\x00\x00commit ")) || (len(line) > 0 && line[len(line)-1] == 0) {
				// Remove trailing null
				if len(line) > 0 && line[len(line)-1] == 0 {
					line = line[:len(line)-1]
				}

				// Split by \x00\x00commit if present
				if idx := bytes.Index(line, []byte("\x00\x00commit ")); idx >= 0 {
					line = line[:idx]
				}

				// Split by null bytes and process files
				start := 0
				for i := 0; i <= len(line); i++ {
					if i == len(line) || line[i] == 0 {
						if i > start {
							// Reuse buffer to avoid allocation
							lineBuffer = append(lineBuffer[:0], line[start:i]...)
							filename := string(lineBuffer)

							if fullpath, exists := filePaths[filename]; exists {
								// Send to worker pool
								updateChan <- fileUpdate{path: fullpath, mtime: commitTime}
								delete(filemap, filename)
								delete(filePaths, filename)
							}
						}
						start = i + 1
					}
				}
			}
		}
	}

	// Close update channel and wait for workers
	close(updateChan)
	wg.Wait()

	// Wait for git command to finish
	if err := cmd.Wait(); err != nil {
		// Ignore error if we stopped early because all files were processed
		if len(filemap) > 0 {
			return fmt.Errorf("git log failed: %w", err)
		}
	}

	// Handle remaining files that weren't found in the commit history
	if len(filemap) > 0 {
		Logger.Warn("Some files not found in commit history", "count", len(filemap))

		// Use the last commit time for remaining files
		for file := range filemap {
			Logger.Warn("File not found", "path", file)
			if fullpath, exists := filePaths[file]; exists {
				os.Chtimes(fullpath, lastCommitTime, lastCommitTime)
			}
		}
	}

	return nil
}

// parseCommitterTimestampFast extracts Unix timestamp from committer line using byte operations
// Expected format: "committer Name <email> timestamp timezone"
func parseCommitterTimestampFast(line []byte) int64 {
	// Find the last two space-separated fields
	lastSpace := bytes.LastIndexByte(line, ' ')
	if lastSpace < 0 {
		return 0
	}

	secondLastSpace := bytes.LastIndexByte(line[:lastSpace], ' ')
	if secondLastSpace < 0 {
		return 0
	}

	// Parse the timestamp (second-to-last field)
	timestampBytes := line[secondLastSpace+1 : lastSpace]
	timestamp, err := strconv.ParseInt(string(timestampBytes), 10, 64)
	if err != nil {
		return 0
	}

	return timestamp
}
