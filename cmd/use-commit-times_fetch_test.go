package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func runTestGit(t *testing.T, dir string, env []string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// newDepth1Clone returns a --depth 1 clone of a local repo with commits
// testfile (2020-06-12 06:03:11 UTC) and other (2021-01-01 UTC).
func newDepth1Clone(t *testing.T) string {
	t.Helper()
	src := t.TempDir()
	runTestGit(t, src, nil, "init", "-q", "-b", "main")
	runTestGit(t, src, nil, "config", "uploadpack.allowFilter", "true")
	runTestGit(t, src, nil, "config", "uploadpack.allowAnySHA1InWant", "true")

	commit := func(name string, date string) {
		if err := os.WriteFile(filepath.Join(src, name), []byte(name), 0644); err != nil {
			t.Fatal(err)
		}
		runTestGit(t, src, nil, "add", name)
		env := []string{
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
			"GIT_AUTHOR_DATE=" + date, "GIT_COMMITTER_DATE=" + date,
		}
		runTestGit(t, src, env, "-c", "commit.gpgsign=false", "commit", "-q", "-m", name)
	}
	commit("testfile", "2020-06-12T06:03:11+00:00")
	commit("other", "2021-01-01T00:00:00+00:00")

	dst := filepath.Join(t.TempDir(), "clone")
	runTestGit(t, "", nil, "clone", "-q", "--depth", "1", "file://"+src, dst)
	return dst
}

func TestUseCommitTimesDepth1(t *testing.T) {
	commitTime, _ := time.Parse(time.RFC3339, "2020-06-12T06:03:11+00:00")
	headTime, _ := time.Parse(time.RFC3339, "2021-01-01T00:00:00+00:00")

	cases := []struct {
		name   string
		fetch  bool
		expect time.Time
	}{
		{"without fetch uses HEAD time", false, headTime},
		{"with fetch uses commit time", true, commitTime},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prevFetch, prevRemote := fetchEnabled, remoteName
			t.Cleanup(func() { fetchEnabled, remoteName = prevFetch, prevRemote })
			fetchEnabled = tc.fetch
			remoteName = "origin"

			dst := newDepth1Clone(t)
			if err := use_commit_times(dst); err != nil {
				t.Fatalf("use_commit_times failed: %v", err)
			}
			info, err := os.Stat(filepath.Join(dst, "testfile"))
			if err != nil {
				t.Fatal(err)
			}
			if !info.ModTime().Equal(tc.expect) {
				t.Fatalf("modtime %s, want %s", info.ModTime().UTC(), tc.expect.UTC())
			}
		})
	}
}

func TestUnshallowRestoresCommitTime(t *testing.T) {
	dst := newDepth1Clone(t)

	if shallow, err := isShallow(dst); err != nil || !shallow {
		t.Fatalf("expected shallow clone: shallow=%v err=%v", shallow, err)
	}
	if err := unshallow(dst, "origin", ""); err != nil {
		t.Fatalf("unshallow failed: %v", err)
	}
	if shallow, err := isShallow(dst); err != nil || shallow {
		t.Fatalf("expected full history after unshallow: shallow=%v err=%v", shallow, err)
	}

	filemap, err := ls_files(dst)
	if err != nil {
		t.Fatal(err)
	}
	if err := use_commit_times_walk(dst, filemap, nil, nil); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(filepath.Join(dst, "testfile"))
	if err != nil {
		t.Fatal(err)
	}
	want, _ := time.Parse(time.RFC3339, "2020-06-12T06:03:11+00:00")
	if !info.ModTime().Equal(want) {
		t.Fatalf("modtime %s, want %s", info.ModTime().UTC(), want.UTC())
	}
}

func TestUnshallowWithTokenOnLocalRemote(t *testing.T) {
	dst := newDepth1Clone(t)
	if err := unshallow(dst, "origin", "dummy-token"); err != nil {
		t.Fatalf("unshallow with token failed: %v", err)
	}
	if shallow, err := isShallow(dst); err != nil || shallow {
		t.Fatalf("expected full history after unshallow: shallow=%v err=%v", shallow, err)
	}
}

func TestCredentialArgsSuppliesTokenOnlyForRemoteHost(t *testing.T) {
	dir := t.TempDir()
	runTestGit(t, dir, nil, "init", "-q")
	runTestGit(t, dir, nil, "remote", "add", "origin", "https://github.com/srz-zumix/example.git")

	args, err := credentialArgs(dir, "origin")
	if err != nil {
		t.Fatalf("credentialArgs failed: %v", err)
	}
	if len(args) == 0 {
		t.Fatal("expected credential args for https remote")
	}

	fill := func(host string) string {
		cmd := exec.Command("git", append(args, "credential", "fill")...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GITHUB_TOKEN=secret-token", "GIT_TERMINAL_PROMPT=0")
		cmd.Stdin = strings.NewReader("protocol=https\nhost=" + host + "\n\n")
		// Exits non-zero when no credential is found (prompting is disabled), which is expected for other hosts.
		out, _ := cmd.Output()
		return string(out)
	}

	if got := fill("github.com"); !strings.Contains(got, "password=secret-token") {
		t.Fatalf("expected token for github.com, got:\n%s", got)
	}
	if got := fill("example.com"); strings.Contains(got, "secret-token") {
		t.Fatalf("token must not be sent to other hosts, got:\n%s", got)
	}
}

func TestCredentialArgsNoneForLocalRemote(t *testing.T) {
	dir := t.TempDir()
	runTestGit(t, dir, nil, "init", "-q")
	runTestGit(t, dir, nil, "remote", "add", "origin", "/some/local/path")

	args, err := credentialArgs(dir, "origin")
	if err != nil {
		t.Fatalf("credentialArgs failed: %v", err)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args for local remote, got %v", args)
	}
}
