package driver

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/blang/semver"
)

type GitTagDriver struct {
	Branch     string
	URI        string
	Repository string
}

const nothingToDescribe = "No names found, cannot describe anything"

func (driver *GitTagDriver) readVersion() (semver.Version, bool, error) {
	var currentVersionStr string

	gitDescribe := exec.Command("git", "describe", "--tags", "--abbrev=0", "origin/"+driver.Branch)
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

func (driver *GitTagDriver) writeVersion(newVersion semver.Version) (bool, error) {
	tagMessage := fmt.Sprintf(
		"Pipeline: %s\nJob: %s\nBuild: %s",
		os.Getenv("BUILD_PIPELINE_NAME"),
		os.Getenv("BUILD_JOB_NAME"),
		os.Getenv("BUILD_NAME"))

	gitTag := exec.Command("git", "tag", "--force", "--annotate", "--message", tagMessage, newVersion.String(), "origin/"+driver.Branch)
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

func (driver *GitTagDriver) setUpRepo() error {

	if driver.Repository != "" {
		_, err := os.Stat(driver.Repository)
		if err == nil {
			return err
		}
		gitRepoDir = driver.Repository
		driver.Branch = "HEAD"
	} else if driver.URI != "" && driver.Branch != "" {
		_, err := os.Stat(gitRepoDir)
		if err != nil {
			gitClone := exec.Command("git", "clone", driver.URI, "--branch", driver.Branch, gitRepoDir)
			gitClone.Stdout = os.Stderr
			gitClone.Stderr = os.Stderr
			if err := gitClone.Run(); err != nil {
				return err
			}
			return nil
		}
	} else {
		return fmt.Errorf("Expected either repository or URI & Branch to be configured.")
	}

	gitFetch := exec.Command("git", "fetch", "origin", driver.Branch, "--tags")
	gitFetch.Dir = gitRepoDir
	gitFetch.Stdout = os.Stderr
	gitFetch.Stderr = os.Stderr
	if err := gitFetch.Run(); err != nil {
		return err
	}

	return nil
}
