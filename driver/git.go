package driver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/mail"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
	"github.com/concourse/semver-resource/version"
)

var gitRepoDir string
var privateKeyPath string
var netRcPath string

var ErrEncryptedKey = errors.New("private keys with passphrases are not supported")

func init() {
	gitRepoDir = filepath.Join(os.TempDir(), "semver-git-repo")
	privateKeyPath = filepath.Join(os.TempDir(), "private-key")
	netRcPath = filepath.Join(os.Getenv("HOME"), ".netrc")
}

type GitDriver struct {
	InitialVersion semver.Version

	URI           string
	Branch        string
	PrivateKey    string
	Username      string
	Password      string
	GitUser       string
	VersionDriver VersionDriver
}

type VersionDriver interface {
	readVersion() (semver.Version, bool, error)
	writeVersion(semver.Version, string) (bool, error)
}

func (driver *GitDriver) Bump(bump version.Bump) (semver.Version, error) {
	err := driver.setUpAuth()
	if err != nil {
		return semver.Version{}, err
	}

	err = driver.setUserInfo()
	if err != nil {
		return semver.Version{}, err
	}

	var newVersion semver.Version

	for {
		err = driver.setUpRepo()
		if err != nil {
			return semver.Version{}, err
		}

		currentVersion, exists, err := driver.VersionDriver.readVersion()
		if err != nil {
			return semver.Version{}, err
		}

		if !exists {
			currentVersion = driver.InitialVersion
		}

		newVersion = bump.Apply(currentVersion)

		wrote, err := driver.VersionDriver.writeVersion(newVersion, driver.Branch)
		if wrote {
			break
		}
	}

	return newVersion, nil
}

func (driver *GitDriver) Set(newVersion semver.Version) error {
	err := driver.setUpAuth()
	if err != nil {
		return err
	}

	err = driver.setUserInfo()
	if err != nil {
		return err
	}

	for {
		err = driver.setUpRepo()
		if err != nil {
			return err
		}

		wrote, err := driver.VersionDriver.writeVersion(newVersion, driver.Branch)
		if err != nil {
			return err
		}

		if wrote {
			break
		}
	}

	return nil
}

func (driver *GitDriver) Check(cursor *semver.Version) ([]semver.Version, error) {
	err := driver.setUpAuth()
	if err != nil {
		return nil, err
	}

	err = driver.setUpRepo()
	if err != nil {
		return nil, err
	}

	currentVersion, exists, err := driver.VersionDriver.readVersion()
	if err != nil {
		return nil, err
	}

	if !exists {
		return []semver.Version{driver.InitialVersion}, nil
	}

	if cursor == nil || currentVersion.GTE(*cursor) {
		return []semver.Version{currentVersion}, nil
	}

	return []semver.Version{}, nil
}

func (driver *GitDriver) setUpRepo() error {

	_, err := os.Stat(gitRepoDir)
	if err != nil {
		gitClone := exec.Command("git", "clone", driver.URI, "--branch", driver.Branch, gitRepoDir)
		gitClone.Stdout = os.Stderr
		gitClone.Stderr = os.Stderr
		if err := gitClone.Run(); err != nil {
			return err
		}
	} else {
		gitFetch := exec.Command("git", "fetch", "origin", driver.Branch, "--tags")
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

func (driver *GitDriver) setUpAuth() error {
	_, err := os.Stat(netRcPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		err := os.Remove(netRcPath)
		if err != nil {
			return err
		}
	}

	if len(driver.PrivateKey) > 0 {
		err := driver.setUpKey()
		if err != nil {
			return err
		}
	}

	if len(driver.Username) > 0 && len(driver.Password) > 0 {
		err := driver.setUpUsernamePassword()
		if err != nil {
			return err
		}
	}

	return nil
}

func (driver *GitDriver) setUpKey() error {
	if strings.Contains(driver.PrivateKey, "ENCRYPTED") {
		return ErrEncryptedKey
	}

	_, err := os.Stat(privateKeyPath)
	if err != nil {
		if os.IsNotExist(err) {
			err := ioutil.WriteFile(privateKeyPath, []byte(driver.PrivateKey), 0600)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return os.Setenv("GIT_SSH_COMMAND", "ssh -o StrictHostKeyChecking=no -i "+privateKeyPath)
}

func (driver *GitDriver) setUpUsernamePassword() error {
	_, err := os.Stat(netRcPath)
	if err != nil {
		if os.IsNotExist(err) {
			content := fmt.Sprintf("default login %s password %s", driver.Username, driver.Password)
			err := ioutil.WriteFile(netRcPath, []byte(content), 0600)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (driver *GitDriver) setUserInfo() error {
	if len(driver.GitUser) == 0 {
		return nil
	}

	e, err := mail.ParseAddress(driver.GitUser)
	if err != nil {
		return err
	}

	if len(e.Name) > 0 {
		gitName := exec.Command("git", "config", "--global", "user.name", e.Name)
		gitName.Stdout = os.Stderr
		gitName.Stderr = os.Stderr
		if err := gitName.Run(); err != nil {
			return err
		}
	}

	gitEmail := exec.Command("git", "config", "--global", "user.email", e.Address)
	gitEmail.Stdout = os.Stderr
	gitEmail.Stderr = os.Stderr
	if err := gitEmail.Run(); err != nil {
		return err
	}
	return nil
}
