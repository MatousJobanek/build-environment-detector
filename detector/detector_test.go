package detector_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/MatousJobanek/build-environment-detector/detector"
	"github.com/MatousJobanek/build-environment-detector/detector/environment"
	"github.com/MatousJobanek/build-environment-detector/detector/git"
	"github.com/stretchr/testify/require"
)

var homeDir = os.Getenv("HOME")

func TestGitHubDetectorWithToken(t *testing.T) {
	token, err := ioutil.ReadFile(homeDir + "/.github-auth")
	require.NoError(t, err)

	ghSource := &git.Source{
		URL:    "https://github.com/wildfly/wildfly",
		Secret: git.NewOauthToken(token),
	}

	buildEnvStats, err := detector.DetectBuildEnvironments(ghSource)
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func TestGitHubDetectorWithUsernameAndPassword(t *testing.T) {
	ghSource := &git.Source{
		URL:    "https://github.com/wildfly/wildfly",
		Secret: git.NewUsernamePassword("anonymous", ""),
	}

	buildEnvStats, err := detector.DetectBuildEnvironments(ghSource)
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func TestGenericGitUsingSshAccessingGitHub(t *testing.T) {

	buffer, err := ioutil.ReadFile(homeDir + "/.ssh/id_rsa")
	require.NoError(t, err)

	ghSource := &git.Source{
		URL:    "git@github.com:wildfly/wildfly.git",
		Secret: git.NewSshKey(buffer, []byte("passphrase")),
	}

	buildEnvStats, err := detector.DetectBuildEnvironments(ghSource)
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func TestGitLabDetectorWithToken(t *testing.T) {

	glSource := &git.Source{
		URL:    "https://gitlab.com/gitlab-org/gitlab-qa",
		Secret: git.NewOauthToken([]byte("")),
	}

	buildEnvStats, err := detector.DetectBuildEnvironments(glSource)
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func TestGenericGitUsingSshAccessingGitLab(t *testing.T) {

	buffer, err := ioutil.ReadFile(homeDir + "/.ssh/id_rsa")
	require.NoError(t, err)

	ghSource := &git.Source{
		URL:    "git@gitlab.cee.redhat.com:mjobanek/housekeeping.git",
		Secret: git.NewSshKey(buffer, []byte("passphrase")),
	}

	buildEnvStats, err := detector.DetectBuildEnvironments(ghSource)
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func printBuildEnvStats(buildEnvStats *environment.BuildEnvStats) {
	fmt.Println(buildEnvStats.SortedLanguages)
	for _, build := range buildEnvStats.DetectedBuildTools {
		fmt.Println(*build)
	}
}
