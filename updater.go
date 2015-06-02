package deploy

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/remind101/deploy/Godeps/_workspace/src/github.com/inconshreveable/go-update"
	"github.com/remind101/deploy/Godeps/_workspace/src/github.com/octokit/go-octokit/octokit"
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
	downloadURL := fmt.Sprintf("https://%s/remind101/deploy/releases/download/%s/deploy_%s_%s_%s.zip", updater.Host, releaseName, version, runtime.GOOS, runtime.GOARCH)

	fmt.Printf("Downloading %s...", version)
	path, err := downloadFile(downloadURL)
	if err != nil {
		return
	}

	exec, err := unzipExecutable(path)
	if err != nil {
		return
	}

	err, _ = update.New().FromFile(exec)
	if err != nil {
		fmt.Printf("Update failed: %v\n", err)
	}

	return
}

func unzipExecutable(path string) (exec string, err error) {
	rc, err := zip.OpenReader(path)
	if err != nil {
		err = fmt.Errorf("Can't open zip file %s: %s", path, err)
		return
	}
	defer rc.Close()

	for _, file := range rc.File {
		if !strings.HasPrefix(file.Name, "deploy") {
			continue
		}

		dir := filepath.Dir(path)
		exec, err = unzipFile(file, dir)
		break
	}

	if exec == "" && err == nil {
		err = fmt.Errorf("No Deploy executable is found in %s", path)
	}

	return
}

func unzipFile(file *zip.File, to string) (exec string, err error) {
	frc, err := file.Open()
	if err != nil {
		err = fmt.Errorf("Can't open zip entry %s when reading: %s", file.Name, err)
		return
	}
	defer frc.Close()

	dest := filepath.Join(to, filepath.Base(file.Name))
	f, err := os.Create(dest)
	if err != nil {
		return
	}
	defer f.Close()

	copied, err := io.Copy(f, frc)
	if err != nil {
		return
	}

	if uint32(copied) != file.UncompressedSize {
		err = fmt.Errorf("Zip entry %s is corrupted", file.Name)
		return
	}

	exec = f.Name()

	return
}

func downloadFile(url string) (path string, err error) {
	dir, err := ioutil.TempDir("", "deploy-update")
	if err != nil {
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		err = fmt.Errorf("Can't download %s: %d", url, resp.StatusCode)
		return
	}

	file, err := os.Create(filepath.Join(dir, filepath.Base(url)))
	if err != nil {
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return
	}

	path = file.Name()

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
