package deploy

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	hub "github.com/github/hub/github"
	"github.com/google/go-github/github"
)

const (
	Name  = "deploy"
	Usage = "A command for creating GitHub deployments"
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name:   "github",
		Value:  "https://api.github.com",
		Usage:  "The location of the GitHub API. You probably don't want to change this.",
		EnvVar: "GITHUB_API_URL",
	},
	cli.StringFlag{
		Name:  "ref",
		Value: "master",
		Usage: "The git ref to deploy. Can be a git commit, branch or tag.",
	},
	cli.StringFlag{
		Name:  "env",
		Value: "",
		Usage: "The environment to deploy to.",
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

	d, _, err := client.Repositories.CreateDeployment(owner, repo, &github.DeploymentRequest{
		Ref:         github.String(c.String("ref")),
		Task:        github.String("deploy"),
		AutoMerge:   github.Bool(false),
		Environment: github.String(c.String("env")),
		// TODO Description:
	})
	if err != nil {
		log.Fatal(err)
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
