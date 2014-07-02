package gogit

import (
	"errors"
	"github.com/jwiklund/tools/debug"
	"os/exec"
	"strings"
)

func RemoteUrl() (string, error) {
	debug.Log("git config --get remote.origin.url")
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	res, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New("Not a git repository, git config got " + err.Error())
	}
	if len(res) == 0 {
		return "", errors.New("Not a git repository, or no origin branch")
	}
	return strings.Trim(string(res), " \t\n\r"), nil
}
