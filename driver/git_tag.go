package driver

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/blang/semver"
)

type GitTagDriver struct {
}

const nothingToDescribe = "No names found, cannot describe anything"

func (*GitTagDriver) readVersion() (semver.Version, bool, error) {
	var currentVersionStr string

	gitDescribe := exec.Command("git", "describe", "--tags", "--abbrev=0")
	gitDescribe.Dir = gitRepoDir
	describeOutput, err := gitDescribe.CombinedOutput()

	currentVersionStr = strings.TrimSpace(string(describeOutput))

	if err != nil {
		if strings.Contains(currentVersionStr, nothingToDescribe) {
			return semver.Version{}, false, nil
		}

		os.Stderr.Write(describeOutput)
		return semver.Version{}, false, err
	}

	currentVersion, err := semver.Parse(currentVersionStr)

	if err != nil {
		return semver.Version{}, false, err
	}

	return currentVersion, true, nil
}

func (*GitTagDriver) writeVersion(newVersion semver.Version, _ string) (bool, error) {
	tagMessage := fmt.Sprintf(
		"Pipeline: %s\nJob: %s\nBuild: %s",
		os.Getenv("BUILD_PIPELINE_NAME"),
		os.Getenv("BUILD_JOB_NAME"),
		os.Getenv("BUILD_NAME"))

	gitTag := exec.Command("git", "tag", "--force", "--annotate", "--message", tagMessage, newVersion.String())
	gitTag.Dir = gitRepoDir
	tagOutput, err := gitTag.CombinedOutput()
	if err != nil {
		os.Stderr.Write(tagOutput)
		return false, err
	}

	gitShowTag := exec.Command("git", "ls-remote", "origin", fmt.Sprintf("refs/tags/%s", newVersion.String()))
	gitShowTag.Dir = gitRepoDir
	showTagOutput, err := gitShowTag.CombinedOutput()

	if err != nil {
		os.Stderr.Write(showTagOutput)
		return false, err
	}

	if strings.Contains(string(showTagOutput), newVersion.String()) {
		gitDeleteTag := exec.Command("git", "push", "origin", fmt.Sprintf(":refs/tags/%s", newVersion.String()))
		gitDeleteTag.Dir = gitRepoDir
		deleteTagOutput, err := gitDeleteTag.CombinedOutput()

		if err != nil {
			os.Stderr.Write(deleteTagOutput)
			return false, err
		}
	}

	gitPushTag := exec.Command("git", "push", "origin", newVersion.String())
	gitPushTag.Dir = gitRepoDir

	pushTagOutput, err := gitPushTag.CombinedOutput()

	if strings.Contains(string(pushTagOutput), pushRejectedString) {
		return false, nil
	}

	if strings.Contains(string(pushTagOutput), pushRemoteRejectedString) {
		return false, nil
	}

	if err != nil {
		os.Stderr.Write(pushTagOutput)
		return false, err
	}

	return true, nil
}
