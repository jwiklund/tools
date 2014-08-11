package main

import (
	"flag"
	"github.com/jwiklund/tools/debug"
	"github.com/jwiklund/tools/gogit"
	"os/exec"
	"strings"
)

func remoteBranch(url string) (string, error) {
	if strings.HasSuffix(url, "spotify-puppet") {
		return "production", nil
	}
	return "master", nil
}

func checkoutRemote(branch string) error {
	debug.Log("git fetch")
	if err := exec.Command("git", "fetch").Run(); err != nil {
		return err
	}

	debug.Log("git checkout local-" + branch)
	output, err := exec.Command("git", "checkout", "local-"+branch).CombinedOutput()
	if err != nil && strings.HasPrefix(string(output), "error: pathspec ") {
		debug.Log("git checkout -t origin/" + branch + " -b local-" + branch)
		err = exec.Command("git", "checkout", "-t", "origin/"+branch, "-b", "local-"+branch).Run()
		if err != nil {
			return err
		}
	} else if err == nil {
		debug.Log("git pull origin master")
		output, err = exec.Command("git", "pull", "origin", branch).CombinedOutput()
		if err != nil {
			debug.Fatalf(string(output))
		}
		debug.Log("git reset origin/" + branch)
		err = exec.Command("git", "reset", "origin/"+branch).Run()
		if err != nil {
			return err
		}
	} else {
		debug.Fatalf(string(output))
	}
	debug.Log("git pull")
	return exec.Command("git", "pull").Run()
}

func main() {
	flag.BoolVar(&debug.Enable, "debug", false, "print debug information")
	flag.Parse()

	url, err := gogit.RemoteUrl()
	if err != nil {
		debug.Fatalf(err.Error())
	}
	debug.Log("origin url " + url)
	remote, err := remoteBranch(url)
	if err != nil {
		debug.Fatalf(err.Error())
	}
	debug.Log("origin branch " + remote)
	err = checkoutRemote(remote)
	if err != nil {
		debug.Fatalf(err.Error())
	}
}
