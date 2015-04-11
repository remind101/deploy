# GitHub Deployments CLI [![Build Status](https://travis-ci.org/remind101/deploy.svg?branch=master)](https://travis-ci.org/remind101/deploy)

A small Go program for creating **[GitHub Deployments](https://developer.github.com/v3/repos/deployments/)**.

## Installation

You can grab the latest release **[here](https://github.com/remind101/deploy/releases)**

Or if you have a working Go 1.4 environment:

```
go get -u github.com/remind101/deploy/cmd/deploy
```

## Usage

The first time you try to deploy, you'll be asked to authenticate with GitHub. If you're already using **[hub](https://github.com/github/hub)**, then you'll already be authenticated.

Deploy the master branch of a repo to staging:

```console
$ deploy --ref=master --env=staging remind101/acme-inc
```

An empty `--ref` flag can mean one of two things:

1. If you're within a git repo, it defaults to the current branch.
2. If you're not within a git repo, then it defaults to `master`.

```console
$ deploy --env=staging remind101/acme-inc
```

You can default to a certain GitHub organization by setting a `GITHUB_ORGANIZATION` environment variable:

```console
$ export GITHUB_ORGANIZATION=remind101
$ deploy --env=staging acme-inc
```

Deploy the current GitHub repo to staging:

```console
$ deploy --env=staging
```

---

Don't have something handling your GitHub Deployment events? Try **[remind101/tugboat](https://github.com/remind101/tugboat)** or **[atmos/heaven](https://github.com/atmos/heaven)**.
