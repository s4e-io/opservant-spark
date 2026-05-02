package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	switch os.Args[1] {
	case "help", "--help", "-h":
		printHelp()
	case "version", "--version", "-v":
		fmt.Println(version)
	case "playbook":
		cmdPlaybook(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}
