# GitHub Deployments CLI

Deploy is a small Go program for creating **[GitHub Deployments](https://developer.github.com/v3/repos/deployments/)**.

## Installation

```
go get -u github.com/remind101/deploy
```

**TODO**: Prebuilt binaries.

## Usage

The first time you try to deploy, you'll be asked to authenticate with GitHub.

Deploy the master branch of a repo to staging:

```console
$ deploy -ref=master -env=staging remind101/r101-api
```

Deploy the current git branch to staging. This requires that you're within a git repo:

```console
$ deploy -env=staging remind101/r101-api
```

Deploy the current GitHub repo to staging:

```console
$ deploy -env=staging
```
