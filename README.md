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

### Shallow clones

For a shallow clone (e.g. `actions/checkout` with the default depth), use `--fetch` to download the commit history from the remote first. Blobs are not downloaded (`--filter=blob:none`), but the repository becomes a partial clone.

```sh
git-use-commit-times --fetch --remote origin
```

```yaml
- uses: actions/checkout@v7
  with:
    persist-credentials: true
- run: git-use-commit-times --fetch
```

## Recommended

The same is possible [git-set-file-times](https://gist.github.com/srz-zumix/0a526e8f9182549cbdb6d880a4477ff0)
