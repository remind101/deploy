# GitHub Deployments CLI

Deploy is a small Go program for creating **[GitHub Deployments](https://developer.github.com/v3/repos/deployments/)**.

## Installation

```
go get -u github.com/remind101/deploy/cmd/deploy
```

**TODO**: Prebuilt binaries.

## Usage

The first time you try to deploy, you'll be asked to authenticate with GitHub. If you already using **[hub](https://github.com/github/hub)**, then you'll already be authenticated.

Deploy the master branch of a repo to staging:

```console
$ deploy -ref=master -env=staging remind101/r101-api
```

An empty `-ref` flag can mean one of two things:

1. If you're within a git repo, it defaults to the current git commit.
2. If you're not within a git repo, then it defaults to `master`.

```console
$ deploy -env=staging remind101/r101-api
```

Deploy the current GitHub repo to staging:

```console
$ deploy -env=staging
```
