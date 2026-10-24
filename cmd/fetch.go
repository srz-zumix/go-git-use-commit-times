package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func runGit(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return out, nil
}

func isShallow(workdir string) (bool, error) {
	out, err := runGit(workdir, "rev-parse", "--is-shallow-repository")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "true", nil
}

// unshallow fetches the full history of HEAD from remote if the repository is shallow.
func unshallow(workdir string, remote string) error {
	shallow, err := isShallow(workdir)
	if err != nil {
		return err
	}
	if !shallow {
		return nil
	}

	out, err := runGit(workdir, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	head := strings.TrimSpace(string(out))

	args := append([]string{"fetch"}, fetchFilterArgs...)
	args = append(args, "--unshallow", "--no-tags", remote, head)
	_, err = runGit(workdir, args...)
	return err
}
