package deploy

import (
	"net/url"
	"testing"

	hub "github.com/remind101/deploy/Godeps/_workspace/src/github.com/github/hub/github"
)

var remotes = map[string]*url.URL{
	"github+git":   parseURL("ssh://git@github.com/remind101/acme-inc.git"),
	"github+https": parseURL("https://github.com/remind101/acme-inc.git"),
	"heroku+git":   parseURL("ssh://git@heroku.com/acme-inc.git"),
}

func TestGitHubRepo(t *testing.T) {
	tests := []struct {
		remotes []hub.Remote
		out     string
	}{
		{[]hub.Remote{{Name: "origin", URL: remotes["github+git"]}}, "remind101/acme-inc"},
		{[]hub.Remote{{Name: "origin", URL: remotes["github+https"]}}, "remind101/acme-inc"},
		{[]hub.Remote{{Name: "origin", URL: remotes["heroku+git"]}}, ""},
	}

	for i, tt := range tests {
		repo := GitHubRepo(tt.remotes)

		if got, want := repo, tt.out; got != want {
			t.Fatalf("#%d: Repo() => %s; want %s", i, got, want)
		}
	}
}

func TestSplitRepo(t *testing.T) {
	tests := []struct {
		in          string
		owner, repo string
		err         error
	}{
		{"remind101/acme-inc", "remind101", "acme-inc", nil},
		{"foo", "", "", errInvalidRepo},
	}

	for _, tt := range tests {
		owner, repo, err := SplitRepo(tt.in)
		if err != tt.err {
			t.Fatalf("err => %v; want %v", err, tt.err)
		}

		if got, want := owner, tt.owner; got != want {
			t.Fatalf("owner => %s; want %s", got, want)
		}

		if got, want := repo, tt.repo; got != want {
			t.Fatalf("repo => %s; want %s", got, want)
		}
	}
}

func parseURL(uri string) *url.URL {
	u, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}
	return u
}
