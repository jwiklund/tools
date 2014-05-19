package main

import (
	"os"
	"os/exec"
	"fmt"
	"errors"
	"strings"
	"flag"
)

var debug = false

func remoteUrl() (string, error) {
	if debug {
		fmt.Println("git config --get remote.origin.url")
	}
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

func remoteBranch(url string) (string, error) {
	if (strings.HasSuffix(url, "spotify-puppet")) {
		return "production", nil
	}
	return "master", nil
}

func checkoutRemote(branch string) error {
	if (debug) {
		fmt.Println("git fetch origin")
	}
	err := exec.Command("git", "fetch", "origin").Run()
	if err != nil {
		return err
	}
	if debug {
		fmt.Println("git checkout local-" + branch)
	}
	err = exec.Command("git", "checkout", "local-" + branch).Run()
	if err != nil {
		if debug {
			fmt.Println("git checkout -t origin/" + branch + " -b local-" + branch)
		}
		err = exec.Command("git", "checkout", "-t", "origin/" + branch, "-b", "local-" + branch).Run()
		if err != nil {
			return err
		}
	}
	if debug {
		fmt.Println("git reset origin/" + branch)
	}
	return exec.Command("git", "reset", "origin/" + branch).Run()
}

func main() {
	flag.BoolVar(&debug, "debug", false, "print debug information")
	flag.Parse()
	url, err := remoteUrl()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if debug {
		fmt.Println("origin url", url)
	}
	remote, err := remoteBranch(url)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if debug {
		fmt.Println("origin branch", remote)
	}
	err = checkoutRemote(remote)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}