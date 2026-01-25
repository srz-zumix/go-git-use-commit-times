//go:build !usegogit

package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5"
)

type FileIdMap = map[string]struct{}

func ls_files(repo *git.Repository) (FileIdMap, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}
	workdir := worktree.Filesystem.Root()

	// Execute git ls-files -z command
	cmd := exec.Command("git", "ls-files", "-z")
	cmd.Dir = workdir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute git ls-files: %w", err)
	}

	// Parse null-terminated output
	files := make(FileIdMap)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Split(scanNullTerminated)

	for scanner.Scan() {
		filename := scanner.Text()
		if filename != "" {
			// Use empty hash as we only need the filename for this implementation
			files[filename] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading git ls-files output: %w", err)
	}

	return files, nil
}

// scanNullTerminated is a split function for bufio.Scanner that splits on null bytes
func scanNullTerminated(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\x00'); i >= 0 {
		return i + 1, data[0:i], nil
	}

	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}

func get_fileidmap(repo *git.Repository, fileList []string) (FileIdMap, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}
	workdir := worktree.Filesystem.Root()

	filemap := make(FileIdMap, len(fileList))

	// Verify each file exists using git ls-files
	for _, path := range fileList {
		cmd := exec.Command("git", "ls-files", "-z", "--", path)
		cmd.Dir = workdir

		output, err := cmd.Output()
		if err != nil {
			Logger.Error("Failed to check file with git ls-files", "path", path, "error", err)
			return nil, fmt.Errorf("failed to check file '%s': %w", path, err)
		}

		// Check if file exists in git
		trimmed := strings.TrimSuffix(string(output), "\x00")
		if trimmed == "" {
			return nil, fmt.Errorf("file '%s' not found in git", path)
		}

		// Use empty hash as we only need the filename for this implementation
		filemap[path] = struct{}{}
	}

	return filemap, nil
}
