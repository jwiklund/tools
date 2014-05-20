package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"
)

var debug = false

var logger = log.New(os.Stderr, "", log.LstdFlags)

func debugf(msg string) {
	if debug {
		logger.Printf(msg)
	}
}

func remoteUrl() (string, error) {
	debugf("git config --get remote.origin.url")
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
	if strings.HasSuffix(url, "spotify-puppet") {
		return "production", nil
	}
	return "master", nil
}

func checkoutRemote(branch string) error {
	debugf("git checkout local-" + branch)
	err := exec.Command("git", "checkout", "local-"+branch).Run()
	if err != nil {
		debugf("git checkout -t origin/" + branch + " -b local-" + branch)
		err = exec.Command("git", "checkout", "-t", "origin/"+branch, "-b", "local-"+branch).Run()
		if err != nil {
			return err
		}
	}
	debugf("git pull")
	return exec.Command("git", "pull").Run()
}

func main() {
	flag.BoolVar(&debug, "debug", false, "print debug information")
	flag.Parse()
	url, err := remoteUrl()
	if err != nil {
		log.Fatalf(err.Error())
	}
	debugf("origin url " + url)
	remote, err := remoteBranch(url)
	if err != nil {
		log.Fatalf(err.Error())
	}
	debugf("origin branch " + remote)
	err = checkoutRemote(remote)
	if err != nil {
		log.Fatalf(err.Error())
	}
}
