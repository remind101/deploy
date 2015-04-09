// package deploy_test containers integration tests for the deploy command.
package deploy_test

import (
	"bytes"
	"strings"
	"testing"

	cli "github.com/remind101/deploy"
)

func TestDeploy(t *testing.T) {
	deploy(t, "-ref=master -env=staging deploycli/acme-inc")
}

// deploy runs a deploy command against a fake github.
func deploy(t testing.TB, command string) (out string) {
	b := new(bytes.Buffer)

	app := cli.NewApp()
	app.Writer = b

	arguments := strings.Split(command, " ")
	arguments = append([]string{"deploy"}, arguments...)
	t.Log(arguments)

	if err := app.Run(arguments); err != nil {
		t.Fatal(err)
	}

	out = b.String()
	t.Log(out)
	return
}
