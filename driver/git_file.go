package driver

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
)

type GitFileDriver struct {
	URI    string
	Branch string
	File   string
}

func (driver *GitFileDriver) readVersion() (semver.Version, bool, error) {
	var currentVersionStr string
	versionFile, err := os.Open(filepath.Join(gitRepoDir, driver.File))
	if err != nil {
		if os.IsNotExist(err) {
			return semver.Version{}, false, nil
		}

		return semver.Version{}, false, err
	}

	defer versionFile.Close()

	_, err = fmt.Fscanf(versionFile, "%s", &currentVersionStr)
	if err != nil {
		return semver.Version{}, false, err
	}

	currentVersion, err := semver.Parse(currentVersionStr)
	if err != nil {
		return semver.Version{}, false, err
	}

	return currentVersion, true, nil
}

const nothingToCommitString = "nothing to commit"
const falsePushString = "Everything up-to-date"
const pushRejectedString = "[rejected]"
const pushRemoteRejectedString = "[remote rejected]"

func (driver *GitFileDriver) writeVersion(newVersion semver.Version) (bool, error) {
	err := ioutil.WriteFile(filepath.Join(gitRepoDir, driver.File), []byte(newVersion.String()+"\n"), 0644)
	if err != nil {
		return false, err
	}

	gitAdd := exec.Command("git", "add", driver.File)
	gitAdd.Dir = gitRepoDir
	gitAdd.Stdout = os.Stderr
	gitAdd.Stderr = os.Stderr
	if err := gitAdd.Run(); err != nil {
		return false, err
	}

	gitCommit := exec.Command("git", "commit", "-m", "bump to "+newVersion.String())
	gitCommit.Dir = gitRepoDir

	commitOutput, err := gitCommit.CombinedOutput()

	if strings.Contains(string(commitOutput), nothingToCommitString) {
		return true, nil
	}

	if err != nil {
		os.Stderr.Write(commitOutput)
		return false, err
	}

	gitPush := exec.Command("git", "push", "origin", "HEAD:"+driver.Branch)
	gitPush.Dir = gitRepoDir

	pushOutput, err := gitPush.CombinedOutput()

	if strings.Contains(string(pushOutput), falsePushString) {
		return false, nil
	}

	if strings.Contains(string(pushOutput), pushRejectedString) {
		return false, nil
	}

	if strings.Contains(string(pushOutput), pushRemoteRejectedString) {
		return false, nil
	}

	if err != nil {
		os.Stderr.Write(pushOutput)
		return false, err
	}

	return true, nil
}

func (driver *GitFileDriver) setUpRepo() error {

	_, err := os.Stat(gitRepoDir)
	if err != nil {
		gitClone := exec.Command("git", "clone", driver.URI, "--branch", driver.Branch, gitRepoDir)
		gitClone.Stdout = os.Stderr
		gitClone.Stderr = os.Stderr
		if err := gitClone.Run(); err != nil {
			return err
		}
	} else {
		gitFetch := exec.Command("git", "fetch", "origin", driver.Branch)
		gitFetch.Dir = gitRepoDir
		gitFetch.Stdout = os.Stderr
		gitFetch.Stderr = os.Stderr
		if err := gitFetch.Run(); err != nil {
			return err
		}
	}

	gitCheckout := exec.Command("git", "reset", "--hard", "origin/"+driver.Branch)
	gitCheckout.Dir = gitRepoDir
	gitCheckout.Stdout = os.Stderr
	gitCheckout.Stderr = os.Stderr
	if err := gitCheckout.Run(); err != nil {
		return err
	}

	return nil
}
