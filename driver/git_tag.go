package driver

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"bytes"

	"github.com/blang/semver"
)

type GitTagDriver struct {
	Prefix     string
	URI        string
	Repository string
}

const nothingToDescribe = "No names found, cannot describe anything"
const notAValidObject = "Not a valid object name HEAD"

func (driver *GitTagDriver) readVersion() (semver.Version, bool, error) {
	gitFetch := exec.Command("git", "fetch", "--tags")
	gitFetch.Dir = gitRepoDir
	gitFetch.Stdout = os.Stderr
	gitFetch.Stderr = os.Stderr
	if err := gitFetch.Run(); err != nil {
		return semver.Version{}, false, err
	}

	gitDescribe := exec.Command("git", "tag", "--sort=-taggerdate", "-l", driver.Prefix+"*")
	gitDescribe.Dir = gitRepoDir
	describeOutput, err := gitDescribe.CombinedOutput()

	var currentVersionStr string
	if index := bytes.Index(describeOutput, []byte{10}); index > 1 {
		currentVersionStr = string(describeOutput[:index])
	} else {
		currentVersionStr = string(describeOutput)
	}
	currentVersionStr = strings.TrimSpace(currentVersionStr)

	if err != nil {
		if strings.Contains(currentVersionStr, nothingToDescribe) ||
			strings.Contains(currentVersionStr, notAValidObject) {
			os.Stderr.Write(describeOutput)
			return semver.Version{}, false, nil
		}

		os.Stderr.Write(describeOutput)
		return semver.Version{}, false, err
	}

	if currentVersionStr == "" {
		return semver.Version{}, false, nil
	}

	currentVersionStr = currentVersionStr[len(driver.Prefix):]
	currentVersion, err := semver.Parse(currentVersionStr)

	if err != nil {
		os.Stderr.Write(describeOutput)
		return semver.Version{}, false, err
	}

	return currentVersion, true, nil
}

func (driver *GitTagDriver) writeVersion(newVersion semver.Version) (bool, error) {
	tagMessage := fmt.Sprintf(
		"\"Pipeline: %s\\nJob: %s\\nBuild: %s\"",
		os.Getenv("BUILD_PIPELINE_NAME"),
		os.Getenv("BUILD_JOB_NAME"),
		os.Getenv("BUILD_NAME"))

	gitFetch := exec.Command("git", "fetch", "--tags", "--dry-run", "--depth=1")
	gitFetch.Dir = gitRepoDir
	gitFetchOutput, err := gitFetch.CombinedOutput()
	if err != nil {
		os.Stderr.Write(gitFetchOutput)
		return false, err
	}

	gitLs := exec.Command("git", "ls-remote", "origin", "HEAD")
	gitLs.Dir = gitRepoDir
	gitLsOutput, err := gitLs.CombinedOutput()
	if err != nil {
		os.Stderr.Write(gitLsOutput)
		return false, err
	}

	headRef := strings.Split(string(gitLsOutput), "\t")[0]

	version := driver.Prefix + newVersion.String()

	gitTag := exec.Command("git", "tag", "--force", "--annotate", "--message", tagMessage, version, headRef)
	gitTag.Dir = gitRepoDir
	tagOutput, err := gitTag.CombinedOutput()
	if err != nil {
		os.Stderr.Write(tagOutput)
		return false, err
	}

	gitShowTag := exec.Command("git", "ls-remote", "origin", fmt.Sprintf("refs/tags/%s", version))
	gitShowTag.Dir = gitRepoDir
	showTagOutput, err := gitShowTag.CombinedOutput()

	if err != nil {
		os.Stderr.Write(showTagOutput)
		return false, err
	}

	if strings.Contains(string(showTagOutput), version) {
		gitDeleteTag := exec.Command("git", "push", "origin", fmt.Sprintf(":refs/tags/%s", version))
		gitDeleteTag.Dir = gitRepoDir
		deleteTagOutput, err := gitDeleteTag.CombinedOutput()

		if err != nil {
			os.Stderr.Write(deleteTagOutput)
			return false, err
		}
	}

	gitPushTag := exec.Command("git", "push", "origin", version)
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
	} else if driver.URI != "" {
		_, err := os.Stat(gitRepoDir)
		if err != nil {
			// Init an empty repo ...
			gitInit := exec.Command("git", "init", gitRepoDir)
			gitInit.Stdout = os.Stderr
			gitInit.Stderr = os.Stderr
			if err := gitInit.Run(); err != nil {
				return err
			}
			// ... and setup the remote
			gitRemote := exec.Command("git", "remote", "add", "origin", driver.URI)
			gitRemote.Dir = gitRepoDir
			gitRemote.Stdout = os.Stderr
			gitRemote.Stderr = os.Stderr
			if err := gitRemote.Run(); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("Expected either repository (path) or URI to be configured.")
	}

	if driver.Prefix == "" {
		driver.Prefix = "v"
	}

	return nil
}
