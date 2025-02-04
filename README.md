
[![GitHub release](https://img.shields.io/github/release/sgaunet/gitlab-token-expiration.svg)](https://github.com/sgaunet/gitlab-token-expiration/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/sgaunet/gitlab-token-expiration)](https://goreportcard.com/report/github.com/sgaunet/gitlab-token-expiration)
![GitHub Downloads](https://img.shields.io/github/downloads/sgaunet/gitlab-token-expiration/total)
[![GoDoc](https://godoc.org/github.com/sgaunet/gitlab-token-expiration?status.svg)](https://godoc.org/github.com/sgaunet/gitlab-token-expiration)
[![License](https://img.shields.io/github/license/sgaunet/gitlab-token-expiration.svg)](LICENSE)

# gitlab-token-expiration

This tool lists all sort of expirable tokens of gitlab projects, gitlab groups and the gitlab personal access token. The purpose of this tool is to give an overview of the expiration date of the tokens.

## Getting started

Example:

```yaml
$ export GITLAB_TOKEN=yourtoken
# export GITLAB_URI=https://your-instance-of-gitlab.com  # optional if you are using another gitlab instance
$ gitlab-token-expiration -h
```

## Development

This project is using :

* Golang
* [task for development](https://taskfile.dev/#/)
* [goreleaser](https://goreleaser.com/)

Use task to compile/create release...

```bash
$ task
task: [default] task -a
task: Available tasks for this project:
* build:            Build the binary
* default:          List tasks
* doc:              Start godoc server
* release:          Create a release
* snapshot:         Create a snapshot release
```
