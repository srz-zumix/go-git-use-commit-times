//go:build !usegogit

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
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

	// Parse git log output more efficiently
	var commitTime time.Time
	var lastCommitTime time.Time
	data := output

	// Collect files to update per commit to reduce syscalls
	filesToUpdate := make([]string, 0, 100)

	// Pre-compile patterns for faster matching
	committerPrefix := []byte("committer ")
	nullNull := []byte("\x00\x00")

	for len(data) > 0 && len(filemap) > 0 {
		// Find next newline
		nl := bytes.IndexByte(data, '\n')
		if nl < 0 {
			break
		}

		line := data[:nl]
		data = data[nl+1:]

		// Fast check for committer line (starts with 'c')
		if len(line) > 10 && line[0] == 'c' && bytes.HasPrefix(line, committerPrefix) {
			// Parse timestamp directly without string conversion
			if ts := parseCommitterTimestampFast(line); ts > 0 {
				commitTime = time.Unix(ts, 0)
				lastCommitTime = commitTime
			}
		} else if len(line) > 0 && (line[len(line)-1] == 0 || bytes.Contains(line, nullNull)) {
			// This line contains file names
			filesToUpdate = filesToUpdate[:0] // Reset slice

			// Remove trailing null
			if line[len(line)-1] == 0 {
				line = line[:len(line)-1]
			}

			// Split by \x00\x00commit if present
			if idx := bytes.Index(line, nullNull); idx >= 0 {
				line = line[:idx]
			}

			// Process null-separated filenames more efficiently
			if len(line) == 0 {
				continue
			}

			// Split by null bytes using bytes.Split for better performance
			parts := bytes.Split(line, []byte{0})
			for _, part := range parts {
				if len(part) == 0 {
					continue
				}
				filename := string(part)
				if _, exists := filemap[filename]; exists {
					filesToUpdate = append(filesToUpdate, filename)
				}
			}

			// Update all files from this commit at once
			if len(filesToUpdate) > 0 {
				for _, file := range filesToUpdate {
					fullpath := filemap[file]
					// Only update if file hasn't been modified or doesn't exist
					if err := os.Chtimes(fullpath, commitTime, commitTime); err == nil {
						delete(filemap, file)
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
