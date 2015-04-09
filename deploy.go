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

const (
	DefaultRef     = "master"
	DefaultTimeout = 20 * time.Second
)

var errTimeout = errors.New("timed out waiting for build to start")

func init() {
	cli.AppHelpTemplate = `USAGE:
   # Deploy the master branch of remind101/acme-inc to staging
   {{.Name}} --env=staging --ref=master remind101/acme-inc

   # Deploy HEAD of the current branch to staging
   {{.Name}} --env=staging remind101/acme-inc

   # Deploy the current GitHub repo to staging
   {{.Name}} --env=staging
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
	cli.BoolFlag{
		Name:  "force, f",
		Usage: "Ignore commit status checks.",
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
				if strings.HasPrefix(err.Message, "Conflict: Commit status checks failed for") {
					msg = "Commit status checks failed. You can bypass commit status checks with the --force flag."
				} else if strings.HasPrefix(err.Message, "No ref found for") {
					msg = fmt.Sprintf("%s. Did you push it to GitHub?", err.Message)
				} else {
					msg = err.Message
				}
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

	fmt.Fprintf(w, "Deploying %s@%s to %s...\n", nwo, *r.Ref, *r.Environment)

	d, _, err := client.Repositories.CreateDeployment(owner, repo, r)
	if err != nil {
		return err
	}

	started := make(chan *github.DeploymentStatus)
	completed := make(chan *github.DeploymentStatus)

	go func() {
		started <- waitState(pendingStates, owner, repo, *d.ID, client)
	}()

	go func() {
		completed <- waitState(completedStates, owner, repo, *d.ID, client)
	}()

	select {
	case <-time.After(DefaultTimeout):
		return errTimeout
	case status := <-started:
		var url string
		if status.TargetURL != nil {
			url = *status.TargetURL
		}
		fmt.Fprintf(w, "%s\n", url)
	}

	status := <-completed

	if isFailed(*status.State) {
		return errors.New("Failed to deploy")
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
		return nil, fmt.Errorf("--env flag is required")
	}

	var contexts *[]string
	if c.Bool("force") {
		s := []string{}
		contexts = &s
	}

	return &github.DeploymentRequest{
		Ref:              github.String(ref),
		Task:             github.String("deploy"),
		AutoMerge:        github.Bool(false),
		Environment:      github.String(env),
		RequiredContexts: contexts,
		Payload: map[string]interface{}{
			"force": c.Bool("force"),
		},
		// TODO Description:
	}, nil
}

var (
	pendingStates   = []string{"pending"}
	completedStates = []string{"success", "error", "failure"}
)

func isFailed(state string) bool {
	return state == "error" || state == "failure"
}

// waitState waits for a deployment status that matches the given states, then
// sends on the returned channel.
func waitState(states []string, owner, repo string, deploymentID int, c *github.Client) *github.DeploymentStatus {
	for {
		<-time.After(1 * time.Second)

		statuses, _, err := c.Repositories.ListDeploymentStatuses(owner, repo, deploymentID, nil)
		if err != nil {
			continue
		}

		status := firstStatus(states, statuses)
		if status != nil {
			return status
		}
	}
}

// firstStatus takes a slice of github.DeploymentStatus and returns the
// first status that matches the provided slice of states.
func firstStatus(states []string, statuses []github.DeploymentStatus) *github.DeploymentStatus {
	for _, ds := range statuses {
		for _, s := range states {
			if ds.State != nil && *ds.State == s {
				return &ds
			}
		}
	}

	return nil
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
