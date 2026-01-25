//go:build !usegogit

package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type FileIdMap = map[string]string

func ls_files(workdir string) (FileIdMap, error) {
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
			files[filename] = filepath.Join(workdir, filename)
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

func get_fileidmap(workdir string, fileList []string) (FileIdMap, error) {
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
		filemap[path] = filepath.Join(workdir, path)
	}

	return filemap, nil
}
