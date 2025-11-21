package driver

import (
	"errors"
	"fmt"
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
var RetriesOnErrorWriteVersion = 3

func init() {
	gitRepoDir = filepath.Join(os.TempDir(), "semver-git-repo")
	privateKeyPath = filepath.Join(os.TempDir(), "private-key")
	netRcPath = filepath.Join(os.Getenv("HOME"), ".netrc")
}

type GitDriver struct {
	InitialVersion semver.Version

	URI                 string
	Branch              string
	PrivateKey          string
	Username            string
	Password            string
	File                string
	GitUser             string
	CommitMessage       string
	SkipSSLVerification bool
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

	for range RetriesOnErrorWriteVersion {
		err = driver.setUpRepo()
		if err != nil {
			return semver.Version{}, err
		}

		currentVersion, exists, err := driver.readVersion()
		if err != nil {
			return semver.Version{}, err
		}

		if !exists {
			currentVersion = driver.InitialVersion
		}

		newVersion = bump.Apply(currentVersion)

		var wrote bool
		wrote, err = driver.writeVersion(newVersion)
		if wrote {
			break
		}
	}
	if err != nil {
		return semver.Version{}, err
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

	for range RetriesOnErrorWriteVersion {
		err = driver.setUpRepo()
		if err != nil {
			return err
		}

		var wrote bool
		wrote, err = driver.writeVersion(newVersion)
		if wrote {
			break
		}
	}
	if err != nil {
		return err
	}

	return nil
}

func (driver *GitDriver) Check(cursor *semver.Version) ([]semver.Version, error) {
	err := driver.setUpAuth()
	if err != nil {
		return nil, err
	}

	err = driver.skipSSLVerificationIfNeeded()
	if err != nil {
		return nil, err
	}

	err = driver.setUpRepo()
	if err != nil {
		return nil, err
	}

	currentVersion, exists, err := driver.readVersion()
	if err != nil {
		return nil, err
	}

	if !exists {
		return []semver.Version{driver.InitialVersion}, nil
	}

	return []semver.Version{currentVersion}, nil
}

func (driver *GitDriver) setUpRepo() error {
	_, err := os.Stat(gitRepoDir)
	if err != nil {
		// Use sparse checkout to only fetch the version file
		gitClone := exec.Command("git", "clone", "--no-checkout", "--filter=blob:none", driver.URI, gitRepoDir)
		gitClone.Stdout = os.Stderr
		gitClone.Stderr = os.Stderr
		if err := gitClone.Run(); err != nil {
			return err
		}

		// Initialize sparse checkout
		gitSparseInit := exec.Command("git", "sparse-checkout", "init", "--cone")
		gitSparseInit.Dir = gitRepoDir
		gitSparseInit.Stdout = os.Stderr
		gitSparseInit.Stderr = os.Stderr
		if err := gitSparseInit.Run(); err != nil {
			return err
		}

		// Set sparse checkout to include only the version file
		gitSparseSet := exec.Command("git", "sparse-checkout", "set", filepath.Dir(driver.File))
		gitSparseSet.Dir = gitRepoDir
		gitSparseSet.Stdout = os.Stderr
		gitSparseSet.Stderr = os.Stderr
		if err := gitSparseSet.Run(); err != nil {
			// If directory is root, set the file directly
			gitSparseSet = exec.Command("git", "sparse-checkout", "set", driver.File)
			gitSparseSet.Dir = gitRepoDir
			gitSparseSet.Stdout = os.Stderr
			gitSparseSet.Stderr = os.Stderr
			if err := gitSparseSet.Run(); err != nil {
				return err
			}
		}

		// Checkout the branch
		gitCheckout := exec.Command("git", "checkout", driver.Branch)
		gitCheckout.Dir = gitRepoDir
		gitCheckout.Stdout = os.Stderr
		gitCheckout.Stderr = os.Stderr
		if err := gitCheckout.Run(); err != nil {
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

func (driver *GitDriver) skipSSLVerificationIfNeeded() error {
	if driver.SkipSSLVerification {
		gitSkipSSLVerification := exec.Command("git", "config", "--global", "http.sslVerify", "false")
		gitSkipSSLVerification.Stdout = os.Stderr
		gitSkipSSLVerification.Stderr = os.Stderr
		if err := gitSkipSSLVerification.Run(); err != nil {
			return err
		}
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
	_, err := os.Stat(privateKeyPath)
	if err != nil {
		if os.IsNotExist(err) {
			privateKey := strings.TrimSuffix(driver.PrivateKey, "\n")
			err := os.WriteFile(privateKeyPath, []byte(privateKey+"\n"), 0600)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if isPrivateKeyEncrypted(privateKeyPath) {
		return ErrEncryptedKey
	}

	return os.Setenv("GIT_SSH_COMMAND", "ssh -o StrictHostKeyChecking=no -i "+privateKeyPath)
}

func isPrivateKeyEncrypted(path string) bool {
	passphrase := ``
	cmd := exec.Command(`ssh-keygen`, `-y`, `-f`, path, `-P`, passphrase)
	err := cmd.Run()

	return err != nil
}

func (driver *GitDriver) setUpUsernamePassword() error {
	_, err := os.Stat(netRcPath)
	if err != nil {
		if os.IsNotExist(err) {
			content := fmt.Sprintf("default login %s password %s", driver.Username, driver.Password)
			err := os.WriteFile(netRcPath, []byte(content), 0600)
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

func (driver *GitDriver) readVersion() (semver.Version, bool, error) {
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

func (driver *GitDriver) writeVersion(newVersion semver.Version) (bool, error) {

	path := filepath.Dir(driver.File)
	if path != "/" && path != "." {
		err := os.MkdirAll(filepath.Join(gitRepoDir, path), 0755)
		if err != nil {
			return false, err
		}
	}

	err := os.WriteFile(filepath.Join(gitRepoDir, driver.File), []byte(newVersion.String()+"\n"), 0644)
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
	var commitMessage string
	if driver.CommitMessage == "" {
		commitMessage = "bump to " + newVersion.String()
	} else {
		commitMessage = strings.ReplaceAll(driver.CommitMessage, "%version%", newVersion.String())
		commitMessage = strings.ReplaceAll(commitMessage, "%file%", driver.File)
	}

	gitCommit := exec.Command("git", "commit", "-m", commitMessage)
	gitCommit.Dir = gitRepoDir

	commitOutput, err := gitCommit.CombinedOutput()

	if strings.Contains(string(commitOutput), nothingToCommitString) {
		os.Stderr.Write([]byte("Nothing to commit, skipping version push\n"))
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
		os.Stderr.Write(pushOutput)
		return true, nil
	}

	if err != nil {
		os.Stderr.Write(pushOutput)
		return false, err
	}

	return true, nil
}
