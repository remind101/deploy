package deploy

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/remind101/deploy/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/remind101/deploy/Godeps/_workspace/src/github.com/github/hub/git"
	hub "github.com/remind101/deploy/Godeps/_workspace/src/github.com/github/hub/github"
	"github.com/remind101/deploy/Godeps/_workspace/src/github.com/google/go-github/github"
)

const (
	Name  = "deploy"
	Usage = "A command for creating GitHub deployments"
)

const DefaultRef = "master"

func init() {
	cli.AppHelpTemplate = `USAGE:
   # Deploy the master branch of remind101/acme-inc to staging
   {{.Name}} -env=staging -ref=master remind101/acme-inc

   # Deploy HEAD of the current branch to staging
   {{.Name}} -env=staging remind101/acme-inc

   # Deploy the current GitHub repo to staging
   {{.Name}} -env=staging
{{if .Flags}}
OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}
`
}

var flags = []cli.Flag{
	cli.StringFlag{
		Name:  "ref",
		Value: "",
		Usage: "The git ref to deploy. Can be a git commit, branch or tag.",
	},
	cli.StringFlag{
		Name:  "env",
		Value: "",
		Usage: "The environment to deploy to.",
	},
	cli.StringFlag{
		Name:   "github",
		Value:  "https://api.github.com",
		Usage:  "The location of the GitHub API. You probably don't want to change this.",
		EnvVar: "GITHUB_API_URL",
	},
}

// NewApp returns a new cli.App for the deploy command.
func NewApp() *cli.App {
	app := cli.NewApp()
	app.Name = Name
	app.Usage = Usage
	app.Flags = flags
	app.Action = RunDeploy

	return app
}

// RunDeploy performs a deploy.
func RunDeploy(c *cli.Context) {
	w := c.App.Writer

	h, err := hub.CurrentConfig().PromptForHost("github.com")
	if err != nil {
		log.Fatal(err)
	}

	client, err := newGitHubClient(c, h)
	if err != nil {
		log.Fatal(err)
	}

	owner, repo := splitRepo(c.Args()[0])

	r, err := newDeploymentRequest(c)
	if err != nil {
		log.Fatal(err)
	}

	d, _, err := client.Repositories.CreateDeployment(owner, repo, r)
	if err != nil {
		msg := err.Error()
		if err, ok := err.(*github.ErrorResponse); ok {
			msg = err.Message
		}

		fmt.Fprintf(os.Stderr, "%s\n", msg)
		os.Exit(-1)
	}

	ch := make(chan *github.DeploymentStatus)

	go func() {
		for {
			statuses, _, err := client.Repositories.ListDeploymentStatuses(owner, repo, *d.ID, nil)
			if err != nil {
				continue
			}

			if len(statuses) != 0 {
				ch <- &statuses[0]
				break
			}
		}
	}()

	timeout := time.Duration(20)
	select {
	case <-time.After(timeout * time.Second):
		fmt.Fprintf(os.Stderr, "No deployment started after waiting %d seconds\n", timeout)
		os.Exit(-1)
	case status := <-ch:
		var url string
		if status.TargetURL != nil {
			url = *status.TargetURL
		}

		fmt.Fprintf(w, "Deployment started: %s\n", url)
	}
}

func newDeploymentRequest(c *cli.Context) (*github.DeploymentRequest, error) {
	ref := c.String("ref")
	if ref == "" {
		r, err := git.Ref("HEAD")
		if err == nil {
			ref = r
		} else {
			ref = DefaultRef
		}
	}

	return &github.DeploymentRequest{
		Ref:         github.String(ref),
		Task:        github.String("deploy"),
		AutoMerge:   github.Bool(false),
		Environment: github.String(c.String("env")),
		// TODO Description:
	}, nil
}

func splitRepo(nwo string) (owner string, repo string) {
	parts := strings.Split(nwo, "/")
	owner = parts[0]
	repo = parts[1]
	return
}

func newGitHubClient(c *cli.Context, h *hub.Host) (*github.Client, error) {
	u, err := url.Parse(c.String("github"))
	if err != nil {
		return nil, err
	}

	t := &transport{
		Username: h.AccessToken,
		Password: "x-oauth-basic",
	}

	client := github.NewClient(&http.Client{Transport: t})
	client.BaseURL = u
	return client, nil
}

type transport struct {
	Username  string
	Password  string
	Transport http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Transport == nil {
		t.Transport = http.DefaultTransport
	}

	req.SetBasicAuth(t.Username, t.Password)

	return t.Transport.RoundTrip(req)
}
