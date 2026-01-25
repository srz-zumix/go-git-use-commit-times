//go:build !usegogit

package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
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

	// Parse null-terminated output directly without Scanner
	filenames := bytes.Split(output, []byte{0})
	files := make(FileIdMap, len(filenames))

	for _, filenameBytes := range filenames {
		if len(filenameBytes) == 0 {
			continue
		}
		filename := string(filenameBytes)
		files[filename] = filepath.Join(workdir, filename)
	}

	return files, nil
}

func get_fileidmap(workdir string, fileList []string) (FileIdMap, error) {
	if len(fileList) == 0 {
		return make(FileIdMap), nil
	}

	// Build arguments: git ls-files -z -- file1 file2 file3 ...
	args := []string{"ls-files", "-z", "--"}
	args = append(args, fileList...)

	cmd := exec.Command("git", args...)
	cmd.Dir = workdir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute git ls-files: %w", err)
	}

	// Parse output to get existing files
	foundFiles := make(map[string]bool)
	for _, filenameBytes := range bytes.Split(output, []byte{0}) {
		if len(filenameBytes) > 0 {
			foundFiles[string(filenameBytes)] = true
		}
	}

	// Verify all requested files were found
	filemap := make(FileIdMap, len(fileList))
	for _, path := range fileList {
		if !foundFiles[path] {
			return nil, fmt.Errorf("file '%s' not found in git", path)
		}
		filemap[path] = filepath.Join(workdir, path)
	}

	return filemap, nil
}
