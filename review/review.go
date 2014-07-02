package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/jwiklund/tools/debug"
	"github.com/jwiklund/tools/gogit"
	"os"
	"os/exec"
	"strings"
)

func remoteBranch(url string) string {
	if strings.HasSuffix(url, "spotify-puppet") {
		return "production"
	}
	return "master"
}

func sliceToString(slices []string) string {
	result := ""
	for _, slice := range slices {
		if len(result) != 0 {
			result = result + " "
		}
		result = result + slice
	}
	return result
}

func reviewers() ([]string, error) {
	r, err := os.Open("/home/jwiklund/reviewer")
	if err != nil {
		return nil, err
	}
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(string(scanner.Text()))
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

func main() {
	flag.BoolVar(&debug.Enable, "debug", false, "print debug information")
	flag.Parse()

	remote, err := gogit.RemoteUrl()
	if err != nil {
		debug.Fatalf(err.Error())
	}

	reviewers, err := reviewers()
	if err != nil {
		debug.Fatalf(err.Error())
	}

	branch := remoteBranch(remote)

	args := []string{}
	args = append(args, "-r")
	for _, reviewer := range reviewers {
		args = append(args, reviewer)
	}
	args = append(args, "-b", branch)
	debug.Log("/usr/local/bin/git-review " + sliceToString(args))

	cmd := exec.Command("/usr/local/bin/git-review", args...)
	output, err := cmd.CombinedOutput()
	fmt.Printf(string(output))
	if err != nil {
		debug.Fatalf(err.Error())
	}
}
