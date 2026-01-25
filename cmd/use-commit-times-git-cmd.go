//go:build !usegogit

package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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

	// Build git log command arguments
	args := []string{"-c", "diff.renames=false", "log", "-m", "-r", "--name-only", "--no-color", "--pretty=raw", "-z"}

	// Add time range if specified
	if since != nil {
		args = append(args, fmt.Sprintf("--since=%s", since.Format(time.RFC3339)))
	}
	if until != nil {
		args = append(args, fmt.Sprintf("--until=%s", until.Format(time.RFC3339)))
	}

	// Execute git log command
	cmd := exec.Command("git", args...)
	cmd.Dir = workdir

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to execute git log: %w", err)
	}

	// Parse git log output
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Split(scanLines)

	var commitTime time.Time
	var lastCommitTime time.Time

	for scanner.Scan() {
		line := scanner.Text()

		// Parse committer line to extract timestamp
		if strings.HasPrefix(line, "committer ") {
			timestamp, err := parseCommitterTimestamp(line)
			if err == nil {
				commitTime = timestamp
				lastCommitTime = timestamp
			}
		} else if strings.Contains(line, "\x00\x00commit ") || strings.HasSuffix(line, "\x00") {
			// Process files that changed in this commit
			processFiles := strings.TrimSuffix(line, "\x00")
			processFiles = strings.Split(processFiles, "\x00\x00commit ")[0]

			if processFiles != "" {
				files := strings.Split(processFiles, "\x00")
				for _, file := range files {
					if file == "" {
						continue
					}

					// Only process files that are in our filemap
					if _, exists := filemap[file]; exists {
						fullpath := filepath.Join(workdir, file)
						err := os.Chtimes(fullpath, commitTime, commitTime)
						if err == nil {
							delete(filemap, file)
						}
					}
				}
			}

			// Stop if all files have been processed
			if len(filemap) == 0 {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading git log output: %w", err)
	}

	// Handle remaining files that weren't found in the commit history
	if len(filemap) > 0 {
		Logger.Warn("Some files not found in commit history", "count", len(filemap))

		// Use the last commit time for remaining files
		for file := range filemap {
			Logger.Warn("File not found", "path", file)
			fullpath := filepath.Join(workdir, file)
			err := os.Chtimes(fullpath, lastCommitTime, lastCommitTime)
			if err != nil {
				Logger.Error("Failed to set time", "file", file, "error", err)
			}
		}
	}

	return nil
}

// parseCommitterTimestamp extracts the Unix timestamp from a committer line
// Expected format: "committer Name <email> timestamp timezone"
func parseCommitterTimestamp(line string) (time.Time, error) {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return time.Time{}, fmt.Errorf("invalid committer line format")
	}

	// The timestamp is the second-to-last field
	timestampStr := parts[len(parts)-2]
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	return time.Unix(timestamp, 0), nil
}

// scanLines is a custom split function for bufio.Scanner that splits on newlines
// but preserves null bytes within lines
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0:i], nil
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}
