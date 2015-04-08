package deploy

import (
	"log"
	"net/url"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/google/go-github/github"
)

const (
	Name  = "deploy"
	Usage = "A command for creating GitHub deployments"
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name:   "github",
		Value:  "",
		Usage:  "The location of the GitHub API. You probably don't want to change this.",
		EnvVar: "GITHUB_API_URL",
	},
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
	client, err := newGitHubClient(c)
	if err != nil {
		log.Fatal(err)
	}

	owner, repo := splitRepo(c.Args()[0])

	_, _, err = client.Repositories.CreateDeployment(owner, repo, &github.DeploymentRequest{
		Ref:         github.String(c.String("ref")),
		Task:        github.String("deploy"),
		AutoMerge:   github.Bool(false),
		Environment: github.String(c.String("env")),
		// TODO Description:
	})
	if err != nil {
		log.Fatal(err)
	}
}

func splitRepo(nwo string) (owner string, repo string) {
	parts := strings.Split(nwo, "/")
	owner = parts[0]
	repo = parts[1]
	return
}

func newGitHubClient(c *cli.Context) (*github.Client, error) {
	u, err := url.Parse(c.String("github"))
	if err != nil {
		return nil, err
	}

	client := github.NewClient(nil)
	client.BaseURL = u
	return client, nil
}
