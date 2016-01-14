package deploy

import (
	"errors"
	"net/url"
	"testing"

	hub "github.com/github/hub/github"
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
		defaultOrg  string
		owner, repo string
		err         error
	}{
		{"remind101/acme-inc", "", "remind101", "acme-inc", nil},
		{"remind101/acme-inc", "foobar", "remind101", "acme-inc", nil},
		{"acme-inc", "remind101", "remind101", "acme-inc", nil},
		{"acme-inc", "", "", "", errInvalidRepo},
		{"", "", "", "", errInvalidRepo},
		{"", "remind101", "", "", errInvalidRepo},
	}

	for _, tt := range tests {
		owner, repo, err := SplitRepo(tt.in, tt.defaultOrg)
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

func TestAliasEnvironment(t *testing.T) {
	tests := []struct {
		env string
		out string
	}{
		{"prod", "production"},
		{"stage", "staging"},
		{"production", "production"},
		{"staging", "staging"},
		{"badenvironment", "badenvironment"},
		{"", ""},
	}
	for i, tt := range tests {
		out := AliasEnvironment(tt.env)

		if got, want := out, tt.out; got != want {
			t.Errorf("#%d: AliasEnvironment => %s; want %s", i, got, want)
		}
	}
}

func TestRef(t *testing.T) {
	tests := []struct {
		ref      string
		headFunc func() (string, error)
		out      string
	}{
		{"master", nil, "master"},
		{"", func() (string, error) { return "", errors.New("no git repo") }, "master"},
		{"", func() (string, error) { return "refs/heads/test-deploy", nil }, "test-deploy"},
	}

	for i, tt := range tests {
		out := Ref(tt.ref, tt.headFunc)

		if got, want := out, tt.out; got != want {
			t.Errorf("#%d: Ref => %s; want %s", i, got, want)
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
