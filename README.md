
[![GitHub release](https://img.shields.io/github/release/sgaunet/gitlab-token-expiration.svg)](https://github.com/sgaunet/gitlab-token-expiration/releases/latest)
![GitHub Downloads](https://img.shields.io/github/downloads/sgaunet/gitlab-token-expiration/total)
![Test Coverage](https://raw.githubusercontent.com/wiki/sgaunet/gitlab-token-expiration/coverage-badge.svg)
[![GoDoc](https://godoc.org/github.com/sgaunet/gitlab-token-expiration?status.svg)](https://godoc.org/github.com/sgaunet/gitlab-token-expiration)
[![Linter](https://github.com/sgaunet/gitlab-token-expiration/actions/workflows/linter.yml/badge.svg)](https://github.com/sgaunet/gitlab-token-expiration/actions/workflows/linter.yml)
[![Release](https://github.com/sgaunet/gitlab-token-expiration/actions/workflows/release.yml/badge.svg)](https://github.com/sgaunet/gitlab-token-expiration/actions/workflows/release.yml)
[![Snapshot](https://github.com/sgaunet/gitlab-token-expiration/actions/workflows/snapshot.yml/badge.svg)](https://github.com/sgaunet/gitlab-token-expiration/actions/workflows/snapshot.yml)
[![License](https://img.shields.io/github/license/sgaunet/gitlab-token-expiration.svg)](LICENSE)

# gitlab-token-expiration

This tool lists all sort of expirable tokens of gitlab projects, gitlab groups and the gitlab personal access token. The purpose of this tool is to give an overview of the expiration date of the tokens.

## Features

* **No ID needed**: `group` and `project` without `--id` discover by themselves every group and project whose tokens you can read, and list all their tokens in one table ([details](#listing-all-groups-and-projects)).
* **Fast on large instances**: tokens are retrieved concurrently, tunable with `-p, --parallel` ([details](#parallelism)).
* **Robust**: a group or project you are not allowed to read (`401`/`403`) is skipped with a warning instead of stopping the whole run.
* **Complete**: every list call is paginated, and `group -i` goes down subgroups at any depth.
* **Filter tokens**: keep or hide tokens by name or by source (group or project path) with regular expressions: `--filter-name`, `--exclude-name`, `--filter-source`, `--exclude-source`, `--ignore-case` ([details](#filtering-tokens)).

## Getting started

```bash
$ export GITLAB_TOKEN=yourtoken  # needs at least the read_api scope
# export GITLAB_URI=https://your-instance-of-gitlab.com  # optional if you are using another gitlab instance

$ gitlab-token-expiration group         # all groups where you are Owner
$ gitlab-token-expiration project       # all projects where you are at least Maintainer
$ gitlab-token-expiration group -p 10   # same, with 10 concurrent API calls

# Only the tokens of the groups under team-a, except the renovate ones
$ gitlab-token-expiration group --filter-source '^team-a/' --exclude-name renovate

$ gitlab-token-expiration -h
```

## Usage

```bash
# Tokens of every group where you have the Owner role
gitlab-token-expiration group

# Tokens of every project where you have at least the Maintainer role
gitlab-token-expiration project

# Tokens of one group, of all its subgroups and of all their projects
gitlab-token-expiration group -i 12345

# Tokens of one group only (no subgroups, no projects)
gitlab-token-expiration group -i 12345 --no-recursive

# Tokens of one project
gitlab-token-expiration project -i 678

# Your personal access tokens
gitlab-token-expiration pat

# Any of the above, restricted to the tokens whose name matches a regex
gitlab-token-expiration project --filter-name '^renovate|-ci$'

# Any of the above, hiding the tokens whose name matches a regex, whatever the case
gitlab-token-expiration pat --exclude-name '^tmp-' --ignore-case
```

### Listing all groups and projects

When `--id` is omitted, the tool discovers by itself the groups and projects whose tokens you are allowed to list:

| Command   | Discovered resources                           | Listed tokens                         |
|-----------|------------------------------------------------|---------------------------------------|
| `group`   | groups where you have the **Owner** role       | group access tokens and deploy tokens |
| `project` | projects where you have at least the **Maintainer** role | project access tokens       |

These are the roles GitLab requires to list the tokens. `GITLAB_TOKEN` needs at least the `read_api` scope.

When several groups or projects are processed (no `--id`, or `group -i` in recursive mode), a group or project for which GitLab denies access (`401` or `403`) is skipped: a warning is printed on stderr and the other ones are still processed. Any other error stops the command. With a single `--id`, an access denied error stops the command.

### Parallelism

Tokens of the groups and projects are retrieved concurrently. The `-p, --parallel` flag sets the maximum number of concurrent GitLab API calls (default: `5`). Lower it if your GitLab instance rate-limits you; requests answered with `429` are retried by the GitLab client.

```bash
gitlab-token-expiration group -p 10
```

### Filtering tokens

Tokens can be kept or hidden by name, or by source (the group or project path shown in the `Source` column):

| Flag                    | Effect                                              | Commands                   |
|-------------------------|-----------------------------------------------------|----------------------------|
| `--filter-name <re>`    | Only show tokens whose name matches                 | `group`, `project`, `pat`  |
| `--exclude-name <re>`   | Hide tokens whose name matches                      | `group`, `project`, `pat`  |
| `--filter-source <re>`  | Only show tokens whose source matches               | `group`, `project`         |
| `--exclude-source <re>` | Hide tokens whose source matches                    | `group`, `project`         |
| `--ignore-case`         | Case-insensitive match, for every pattern           | `group`, `project`, `pat`  |

A token is shown when it matches every `--filter-*` pattern given and none of the `--exclude-*` patterns. Patterns are regular expressions ([Go RE2 syntax](https://github.com/google/re2/wiki/Syntax)) matched anywhere in the value: use `|` to match several values, `^` and `$` to anchor. The filters are applied after the tokens are retrieved, so they do not reduce the number of API calls.

```bash
# Only the tokens of the groups and projects under team-a
gitlab-token-expiration group --filter-source '^team-a/'

# Same, except the renovate tokens
gitlab-token-expiration group --filter-source '^team-a/' --exclude-name renovate

# Every token except those of the sandbox groups and projects
gitlab-token-expiration group --exclude-source '(^|/)sandbox(/|$)'

# Only the CI tokens, whatever the case (ci-deploy, Deploy-CI, ...)
gitlab-token-expiration project --filter-name '(^|-)ci(-|$)' --ignore-case
```

### Common flags

| Flag                                 | Description                                                    |
|--------------------------------------|----------------------------------------------------------------|
| `-i, --id`                           | Group or project ID (`group` and `project` only)               |
| `-p, --parallel`                     | Maximum number of concurrent GitLab API calls (default `5`, `group` and `project` only) |
| `-d, --days-before-expiration`       | Tokens expiring within this number of days are shown in yellow (default `60`) |
| `--filter-name`, `--exclude-name`    | Only show / hide tokens whose name matches a regular expression ([details](#filtering-tokens)) |
| `--filter-source`, `--exclude-source` | Only show / hide tokens whose source matches a regular expression (`group` and `project` only) |
| `--ignore-case`                      | Case-insensitive match for every filter pattern                |
| `-r, --revoked`                      | Also print revoked tokens                                      |
| `-H, --no-header`                    | Do not print the table header                                  |
| `-C, --no-color`                     | Do not print colors                                            |

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

## Installation

### From releases

Download the latest release from the [release page](https://github.com/sgaunet/gitlab-token-expiration/releases) and install it in your PATH.

### Homebrew

```bash
brew tap sgaunet/homebrew-tools
brew install sgaunet/tools/gitlab-token-expiration
```
