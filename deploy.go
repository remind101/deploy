package deploy

import (
	"errors"
	"fmt"
	"os"
	"regexp"
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
		Name:  "ref, branch, commit, tag",
		Value: "",
		Usage: "The git ref to deploy. Can be a git commit, branch or tag.",
	},
	cli.StringFlag{
		Name:  "env, e",
		Value: "",
		Usage: "The environment to deploy to.",
	},
}

// NewApp returns a new cli.App for the deploy command.
func NewApp() *cli.App {
	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Name = Name
	app.Usage = Usage
	app.Flags = flags
	app.Action = func(c *cli.Context) {
		if err := RunDeploy(c); err != nil {
			msg := err.Error()
			if err, ok := err.(*github.ErrorResponse); ok {
				msg = err.Message
			}

			fmt.Println(msg)
			os.Exit(-1)
		}
	}

	return app
}

// RunDeploy performs a deploy.
func RunDeploy(c *cli.Context) error {
	w := c.App.Writer

	h, err := hub.CurrentConfig().PromptForHost("github.com")
	if err != nil {
		return err
	}

	client, err := newGitHubClient(h)
	if err != nil {
		return err
	}

	nwo, err := Repo(c.Args())
	if err != nil {
		return err
	}

	owner, repo, err := SplitRepo(nwo)
	if err != nil {
		return fmt.Errorf("Invalid GitHub repo: %s", nwo)
	}

	r, err := newDeploymentRequest(c)
	if err != nil {
		return err
	}

	d, _, err := client.Repositories.CreateDeployment(owner, repo, r)
	if err != nil {
		return err
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
		return fmt.Errorf("No deployment started after waiting %d seconds\n", timeout)
	case status := <-ch:
		var url string
		if status.TargetURL != nil {
			url = *status.TargetURL
		}

		fmt.Fprintf(w, "%s\n", url)
	}

	return nil
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

	env := c.String("env")
	if env == "" {
		return nil, fmt.Errorf("-env flag is required")
	}

	return &github.DeploymentRequest{
		Ref:         github.String(ref),
		Task:        github.String("deploy"),
		AutoMerge:   github.Bool(false),
		Environment: github.String(env),
		// TODO Description:
	}, nil
}

// Repo will determine the correct GitHub repo to deploy to, based on a set of
// arguments.
func Repo(arguments []string) (string, error) {
	if len(arguments) != 0 {
		return arguments[0], nil
	}

	remotes, err := hub.Remotes()
	if err != nil {
		return "", err
	}

	repo := GitHubRepo(remotes)
	if repo == "" {
		return repo, errors.New("no GitHub repo found in .git/config")
	}

	return repo, nil
}

// A regular expression that can convert a URL.Path into a GitHub repo name.
var remoteRegex = regexp.MustCompile(`^/(.*)\.git$`)

// GitHubRepo, given a list of git remotes, will determine what the GitHub repo
// is.
func GitHubRepo(remotes []hub.Remote) string {
	// We only want to look at the `origin` remote.
	remote := findRemote("origin", remotes)
	if remote == nil {
		return ""
	}

	// Remotes that are not pointed at a GitHub repo are not valid.
	if remote.URL.Host != "github.com" {
		return ""
	}

	// Convert `/remind101/acme-inc.git` => `remind101/acme-inc`.
	return remoteRegex.ReplaceAllString(remote.URL.Path, "$1")
}

func findRemote(name string, remotes []hub.Remote) *hub.Remote {
	for _, r := range remotes {
		if r.Name == name {
			return &r
		}
	}

	return nil
}

var errInvalidRepo = errors.New("invalid repo")

// SplitRepo splits a repo string in the form remind101/acme-inc into it's owner
// and repo components.
func SplitRepo(nwo string) (owner string, repo string, err error) {
	parts := strings.Split(nwo, "/")

	if len(parts) != 2 {
		err = errInvalidRepo
		return
	}

	owner = parts[0]
	repo = parts[1]

	return
}
