package main

import "os"

func main() {
	if len(os.Args) < 3 {
		os.Exit(1)
	}
	length := len(os.Args[1])
	if length > len(os.Args[2]) {
		length = len(os.Args[2])
	}
	if os.Args[2][0:length] == os.Args[1] {
		os.Exit(0)
	}
	os.Exit(1)
}
