package main

import (
	"strings"
)

func getNamedArg(args []string, name string) string {
	prefix := "-" + name + "="
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, prefix) {
			return arg[len(prefix):]
		}
	}
	return ""
}

func parseArgs(args []string) (subcommand string, help bool, version bool) {
	// Returns the important details of the CLI arguments.

	for _, arg := range args[1:] {
		if subcommand == "" && len(arg) > 0 && arg[0:1] != "-" {
			subcommand = arg
			break
		} else if arg == "-help" {
			help = true
		} else if arg == "-version" {
			version = true
		}
	}
	return subcommand, help, version
}
