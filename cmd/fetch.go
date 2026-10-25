package cmd

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

// tokenCredentialHelper answers git's credential requests with the token from $GITHUB_TOKEN.
// The token is read from the environment so it never appears in the command line.
const tokenCredentialHelper = `!f() { test "$1" = get && printf 'username=x-access-token\npassword=%s\n' "$GITHUB_TOKEN"; }; f`

func runGit(dir string, args ...string) ([]byte, error) {
	return runGitWithEnv(dir, nil, args...)
}

func runGitWithEnv(dir string, env []string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(append(os.Environ(), "GIT_TERMINAL_PROMPT=0"), env...)
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

// credentialArgs returns git "-c" options that make the remote's HTTP(S) host use the token
// from $GITHUB_TOKEN. The setting is scoped to the remote's scheme and host so the token is
// not offered to other hosts. Non-HTTP(S) remotes get no options.
func credentialArgs(workdir string, remote string) ([]string, error) {
	out, err := runGit(workdir, "remote", "get-url", remote)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(strings.TrimSpace(string(out)))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, nil
	}
	prefix := fmt.Sprintf("credential.%s://%s.", u.Scheme, u.Host)
	// The empty helper resets helpers from other configs so the token takes precedence.
	return []string{
		"-c", prefix + "helper=",
		"-c", prefix + "helper=" + tokenCredentialHelper,
	}, nil
}

// unshallow fetches the full history of HEAD from remote if the repository is shallow.
// If token is non-empty, it is used for HTTP(S) authentication to the remote.
func unshallow(workdir string, remote string, token string) error {
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

	var cfgArgs []string
	var env []string
	if token != "" {
		cfgArgs, err = credentialArgs(workdir, remote)
		if err != nil {
			return err
		}
		env = []string{"GITHUB_TOKEN=" + token}
	}

	args := append(cfgArgs, "fetch")
	args = append(args, fetchFilterArgs...)
	args = append(args, "--unshallow", "--no-tags", remote, head)
	_, err = runGitWithEnv(workdir, env, args...)
	return err
}
