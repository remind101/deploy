// package deploy_test containers integration tests for the deploy command.
package deploy_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
	cli "github.com/remind101/deploy"
)

func TestDeploy(t *testing.T) {
	g := newGH()
	deploy(t, g, "-ref=develop -env=staging remind101/acme-inc")

	if got, want := len(g.deployments["remind101/acme-inc"]), 1; got != want {
		t.Fatal("expected a deployment to be created")
	}

	d := g.deployments["remind101/acme-inc"][0]

	if got, want := *d.Ref, "develop"; got != want {
		t.Fatalf("Ref => %s; want %s", got, want)
	}
}

// deploy runs a deploy command against a fake github.
func deploy(t testing.TB, g *gh, command string) (out string) {
	b := new(bytes.Buffer)

	app := cli.NewApp()
	app.Writer = b

	s := newGitHubServer(g)
	defer s.Close()

	arguments := strings.Split(command, " ")
	arguments = append([]string{"deploy", fmt.Sprintf("-github=%s", s.URL)}, arguments...)
	t.Log(arguments)

	if err := app.Run(arguments); err != nil {
		t.Fatal(err)
	}

	out = b.String()
	t.Log(out)
	return
}

// newGitHubServer returns a new httptest.Server that serves a fake github api.
func newGitHubServer(g *gh) *httptest.Server {
	m := mux.NewRouter()
	m.HandleFunc("/repos/{owner}/{repo}/deployments", g.postDeployments).Methods("POST")

	s := httptest.NewServer(m)
	return s
}

type gh struct {
	// auto incremented deployment id.
	id int

	// created deployments.
	deployments map[string][]*github.Deployment
}

func newGH() *gh {
	return &gh{
		deployments: make(map[string][]*github.Deployment),
	}
}

func (g *gh) postDeployments(w http.ResponseWriter, r *http.Request) {
	var f github.DeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	g.id++
	d := &github.Deployment{
		ID:          github.Int(g.id),
		SHA:         github.String("abcd"),
		Ref:         f.Ref,
		Task:        f.Task,
		Environment: f.Environment,
		Description: f.Description,
	}
	vars := mux.Vars(r)

	key := fmt.Sprintf("%s/%s", vars["owner"], vars["repo"])
	g.deployments[key] = append(g.deployments[key], d)

	json.NewEncoder(w).Encode(d)
}
