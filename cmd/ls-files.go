package cmd

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type FileIdMap = map[string]plumbing.Hash

func ls_files(repo *git.Repository) (FileIdMap, error) {
	Logger.Debug("Listing all files in repository")
	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	files := make(FileIdMap)
	err = tree.Files().ForEach(func(f *object.File) error {
		files[f.Name] = f.Hash
		return nil
	})
	if err != nil {
		Logger.Warn("Failed to list files with go-git, using fallback", "error", err)
		return nil, err
	}
	Logger.Debug("Listed files successfully", "count", len(files))
	return files, nil
}

func get_fileidmap(repo *git.Repository, fileList []string) (FileIdMap, error) {
	Logger.Debug("Getting file IDs", "count", len(fileList))
	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	filemap := make(FileIdMap, len(fileList))
	for _, path := range fileList {
		file, err := tree.File(path)
		if err != nil {
			Logger.Error("Failed to find file in tree", "path", path, "tree", tree.Hash, "error", err)
			return nil, fmt.Errorf("failed to find file '%s' in tree %s: %w", path, tree.Hash, err)
		}
		filemap[path] = file.Hash
	}
	Logger.Debug("Got file IDs successfully", "count", len(filemap))
	return filemap, nil
}
