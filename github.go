package deploy

import (
	"net/http"

	hub "github.com/github/hub/github"
	"github.com/google/go-github/v35/github"
)

// newGitHubClient returns a new github.Client configured for the GitHub Host.
func newGitHubClient(h *hub.Host) (*github.Client, error) {
	t := &transport{
		Token: h.AccessToken,
	}

	client := github.NewClient(&http.Client{Transport: t})
	return client, nil
}

// transport is an http.RoundTripper that adds a GitHub auth token as the basic
// auth credentials.
type transport struct {
	Token     string
	Transport http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Transport == nil {
		t.Transport = http.DefaultTransport
	}

	req.SetBasicAuth(t.Token, "x-oauth-basic")
	return t.Transport.RoundTrip(req)
}
