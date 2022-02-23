package main

import (
	"strings"
)

// cleanArgs converts `-var value` and `-var-file value` arguments
// into `-var=value` and `-var-file=value` respectively.
func cleanArgs(args []string) []string {
	result := []string{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if (arg == "-var" || arg == "-var-file") && i < len(args)-1 {
			result = append(result, arg+"="+args[i+1])
			i = i + 1
		} else {
			result = append(result, arg)
		}
	}
	return result
}

func getNamedArg(args []string, name string) string {
	prefix := "-" + name + "="
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, prefix) {
			return arg[len(prefix):]
		}
	}
	return ""
}

// parseArgs checks CLI arguments and environment variables
// containing extra CLI arguments and returns useful details.
func parseArgs(args []string, env []string) (subcommand string, help bool, version bool, err error) {

	args, err = getArgsWithEnv(args, env)
	if err != nil {
		return "", false, false, err
	}

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

	return subcommand, help, version, nil
}
