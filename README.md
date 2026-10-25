# git-use-commit-times

Set files mtime by lates commit time.

## Install

```sh
brew install srz-zumix/tap/go-git-use-commit-times
```

## Usage

```sh
git-use-commit-times
```

### Author date

By default the committer date is used. Use `--author-date` to set the author date instead.

```sh
git-use-commit-times --author-date
```

### Shallow clones

For a shallow clone (e.g. `actions/checkout` with the default depth), use `--fetch` to download the commit history from the remote first. In the default build, blobs are not downloaded (`--filter=blob:none`), but the repository becomes a partial clone. In the `usegogit` build, blobs are also downloaded because go-git does not support partial clones.

```sh
git-use-commit-times --fetch --remote origin
```

To authenticate the fetch to an HTTP(S) remote (e.g. a private repository), set the `GITHUB_TOKEN` environment variable or pass `--github-token`. The token is supplied only to the remote's scheme and host. Prefer the environment variable, since a command-line value can be visible to other users in the process list.

```sh
GITHUB_TOKEN=... git-use-commit-times --fetch
```

```yaml
- uses: actions/checkout@v7
  with:
    persist-credentials: false
- run: git-use-commit-times --fetch
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Recommended

The same is possible [git-set-file-times](https://gist.github.com/srz-zumix/0a526e8f9182549cbdb6d880a4477ff0)
