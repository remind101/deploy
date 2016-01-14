package deploy

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/inconshreveable/go-update"
	"github.com/octokit/go-octokit/octokit"
)

const GitHubHost = "github.com"

type Updater struct {
	Host           string
	CurrentVersion string
}

func NewUpdater() *Updater {
	version := Version
	return &Updater{
		Host:           GitHubHost,
		CurrentVersion: version,
	}
}

func (updater *Updater) Update() (err error) {
	releaseName, version := updater.latestReleaseNameAndVersion()

	if version == "" {
		fmt.Println("There is no newer version of Deploy available.")
		return
	}

	if version == updater.CurrentVersion {
		fmt.Printf("You're already on the latest version: %s\n", version)
	} else {
		err = updater.updateTo(releaseName, version)
	}

	return
}

func (updater *Updater) updateTo(releaseName, version string) (err error) {
	downloadURL := fmt.Sprintf("https://%s/remind101/deploy/releases/download/%s/%s_%s_deploy", updater.Host, releaseName, runtime.GOOS, runtime.GOARCH)

	fmt.Printf("Downloading %s...", version)
	err, _ = update.New().FromUrl(downloadURL)
	if err != nil {
		fmt.Printf("Update failed: %v\n", err)
	}

	return
}

func (updater *Updater) latestReleaseNameAndVersion() (name, version string) {
	client := octokit.NewClient(nil)

	url, err := octokit.ReleasesURL.Expand(octokit.M{"owner": "remind101", "repo": "deploy"})
	if err != nil {
		return
	}

	releases, result := client.Releases(url).All()
	if result.HasError() {
		err = fmt.Errorf("Error getting Deploy release: %s", result.Err)
		return
	}

	if len(releases) == 0 {
		err = fmt.Errorf("No Deploy release is available")
		return
	}

	name = releases[0].TagName
	version = strings.TrimPrefix(name, "v")
	return
}
