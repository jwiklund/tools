package main

import (
	"flag"
	"github.com/jwiklund/tools/debug"
	"os/exec"
	"strings"
)

func remote() (string, error) {
	debug.Log("git remote")
	output, err := exec.Command("git", "remote").CombinedOutput()
	if err != nil {
		return "", err
	}
	if strings.Contains(string(output), "upstream\n") {
		return "upstream", nil
	}
	return "origin", nil
}

func anyOf(url string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(url, suffix) {
			return true
		}
	}
	return false
}

func checkoutRemote(branch, remote string) error {
	debug.Log("git fetch " + remote)
	if err := exec.Command("git", "fetch", remote).Run(); err != nil {
		return err
	}

	debug.Log("git checkout local-" + branch)
	output, err := exec.Command("git", "checkout", "local-"+branch).CombinedOutput()
	if err != nil && strings.HasPrefix(string(output), "error: pathspec ") {
		debug.Log("git checkout -t " + remote + "/" + branch + " -b local-" + branch)
		err = exec.Command("git", "checkout", "-t", remote+"/"+branch, "-b", "local-"+branch).Run()
		if err != nil {
			return err
		}
	} else if err == nil {
		debug.Log("git merge " + remote + "/" + branch)
		err = exec.Command("git", "merge", remote+"/"+branch).Run()
		if err != nil {
			debug.Log("merge failed (" + err.Error() + "), ignoring")
		}
		debug.Log("git reset " + remote + "/" + branch)
		err = exec.Command("git", "reset", remote+"/"+branch).Run()
		if err != nil {
			return err
		}
	} else {
		debug.Fatalf(string(output))
	}
	return err
}

func main() {
	flag.BoolVar(&debug.Enable, "debug", false, "print debug information")
	flag.Parse()

	remote, err := remote()
	if err != nil {
		debug.Fatalf(err.Error())
	}
	debug.Log("remote " + remote)
	err = checkoutRemote("master", remote)
	if err != nil {
		debug.Fatalf(err.Error())
	}
}
