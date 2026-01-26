//go:build !usegogit

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// use_commit_times_git_cmd implements the same logic as the Perl script
// https://gist.github.com/srz-zumix/0a526e8f9182549cbdb6d880a4477ff0
// using git command line instead of go-git library
func use_commit_times_walk(workdir string, filemap FileIdMap, since *time.Time, until *time.Time) error {
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

	// Parse git log output following the Perl script logic
	var commitTime time.Time
	var lastCommitTime time.Time

	// Split output by newlines (Perl reads line by line with $/ = "\n")
	lines := bytes.Split(output, []byte{'\n'})

	for _, line := range lines {
		if len(filemap) == 0 {
			break
		}

		// Parse committer line to extract timestamp
		if bytes.HasPrefix(line, []byte("committer ")) {
			if ts := parseCommitterTimestampFast(line); ts > 0 {
				commitTime = time.Unix(ts, 0)
				lastCommitTime = commitTime
			}
		} else if bytes.Contains(line, []byte{0, 0, 'c', 'o', 'm', 'm', 'i', 't', ' '}) || bytes.HasSuffix(line, []byte{0}) {
			// This line contains file names
			// Remove the pattern "\0\0commit [hash]..." if present
			if idx := bytes.Index(line, []byte{0, 0, 'c', 'o', 'm', 'm', 'i', 't', ' '}); idx >= 0 {
				line = line[:idx]
			} else if bytes.HasSuffix(line, []byte{0}) {
				// Remove trailing null
				line = line[:len(line)-1]
			}

			if len(line) == 0 {
				continue
			}

			// Split by null bytes to get file names
			files := bytes.Split(line, []byte{0})
			for _, fileBytes := range files {
				if len(fileBytes) == 0 {
					continue
				}

				filename := string(fileBytes)
				if fullpath, exists := filemap[filename]; exists {
					if os.Chtimes(fullpath, commitTime, commitTime) == nil {
						delete(filemap, filename)
					}
				}
			}
		}
	}

	// Handle remaining files that weren't found in the commit history
	if len(filemap) > 0 {
		Logger.Warn("Some files not found in commit history", "count", len(filemap))

		// Use the last commit time for remaining files
		for file := range filemap {
			Logger.Warn("File not found", "path", file)
			if fullpath, exists := filemap[file]; exists {
				err := os.Chtimes(fullpath, lastCommitTime, lastCommitTime)
				if err != nil {
					Logger.Error("Failed to update timestamp for file", "path", fullpath, "error", err)
				}
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
	if lastSpace < 10 {
		return 0
	}

	secondLastSpace := bytes.LastIndexByte(line[:lastSpace], ' ')
	if secondLastSpace < 0 {
		return 0
	}

	// Parse the timestamp manually (faster than strconv.ParseInt for this case)
	var timestamp int64
	for i := secondLastSpace + 1; i < lastSpace; i++ {
		if line[i] < '0' || line[i] > '9' {
			return 0
		}
		timestamp = timestamp*10 + int64(line[i]-'0')
	}

	return timestamp
}
