package main

import (
	"testing"

	"github.com/matryer/is"
)

func testArgs(t *testing.T, args []string, env []string, expected arguments) {
	is := is.New(t)

	// Already arranged

	// Act

	got, err := newArguments(args, env)
	is.NoErr(err)

	// Assert

	is.Equal(got.bin, expected.bin)
	is.Equal(got.cli, expected.cli)
	is.Equal(got.virtual, expected.virtual)
	is.Equal(got.chdir, expected.chdir)
	is.Equal(got.subcommand, expected.subcommand)
	is.Equal(got.help, expected.help)
	is.Equal(got.version, expected.version)
}

func TestArgumentsChdir(t *testing.T) {
	testArgs(t, []string{"ltf", "-chdir=..", "plan"}, []string{}, arguments{
		bin:        "ltf",
		cli:        []string{"ltf", "-chdir=..", "plan"},
		virtual:    []string{"ltf", "-chdir=..", "plan"},
		chdir:      "..",
		subcommand: "plan",
	})
}

func TestArgumentsHelp(t *testing.T) {
	t.Run("flag", func(t *testing.T) {
		testArgs(t, []string{"ltf", "-help"}, []string{}, arguments{
			bin:     "ltf",
			cli:     []string{"ltf", "-help"},
			virtual: []string{"ltf", "-help"},
			help:    true, // the flag should work
		})
	})

	t.Run("subcommand", func(t *testing.T) {
		testArgs(t, []string{"ltf", "help"}, []string{}, arguments{
			bin:        "ltf",
			cli:        []string{"ltf", "help"},
			virtual:    []string{"ltf", "help"},
			subcommand: "help",
			help:       false, // the subcommand is not correct usage
		})
	})
}

func TestArgumentsEmpty(t *testing.T) {
	testArgs(t, []string{"ltf"}, []string{}, arguments{
		bin:     "ltf",
		cli:     []string{"ltf"},
		virtual: []string{"ltf"},
	})
}

func TestArgumentsVars(t *testing.T) {
	t.Run("combined var arg", func(t *testing.T) {
		testArgs(t, []string{"ltf", "plan", "-var=one=1"}, []string{}, arguments{
			bin:        "ltf",
			cli:        []string{"ltf", "plan", "-var=one=1"},
			virtual:    []string{"ltf", "plan", "-var=one=1"},
			subcommand: "plan",
		})
	})

	t.Run("separate var args", func(t *testing.T) {
		testArgs(t, []string{"ltf", "plan", "-var", "one=1"}, []string{}, arguments{
			bin:        "ltf",
			cli:        []string{"ltf", "plan", "-var", "one=1"},
			virtual:    []string{"ltf", "plan", "-var=one=1"},
			subcommand: "plan",
		})
	})

	t.Run("combined var-file arg", func(t *testing.T) {
		testArgs(t, []string{"ltf", "plan", "-var-file=test.tfvars"}, []string{}, arguments{
			bin:        "ltf",
			cli:        []string{"ltf", "plan", "-var-file=test.tfvars"},
			virtual:    []string{"ltf", "plan", "-var-file=test.tfvars"},
			subcommand: "plan",
		})
	})

	t.Run("separate var-file args", func(t *testing.T) {
		testArgs(t, []string{"ltf", "plan", "-var-file", "test.tfvars"}, []string{}, arguments{
			bin:        "ltf",
			cli:        []string{"ltf", "plan", "-var-file", "test.tfvars"},
			virtual:    []string{"ltf", "plan", "-var-file=test.tfvars"},
			subcommand: "plan",
		})
	})

	// From the Terraform docs:
	// These arguments are inserted directly after the subcommand (such as plan)
	// and before any flags specified directly on the command-line. This behavior
	// ensures that flags on the command-line take precedence over environment variables.

	t.Run("env args", func(t *testing.T) {
		testArgs(t, []string{"ltf", "plan", "-var", "one=1"}, []string{"TF_CLI_ARGS=-var=two=2 -var three=3"}, arguments{
			bin:        "ltf",
			cli:        []string{"ltf", "plan", "-var", "one=1"},
			virtual:    []string{"ltf", "plan", "-var=two=2", "-var=three=3", "-var=one=1"},
			subcommand: "plan",
		})
	})

	t.Run("subcommand env args", func(t *testing.T) {
		testArgs(t, []string{"ltf", "plan", "-var", "one=1"}, []string{"TF_CLI_ARGS_plan=-var=two=2"}, arguments{
			bin:        "ltf",
			cli:        []string{"ltf", "plan", "-var", "one=1"},
			virtual:    []string{"ltf", "plan", "-var=two=2", "-var=one=1"},
			subcommand: "plan",
		})
	})

	t.Run("wrong subcommand env args", func(t *testing.T) {
		// When doing a "plan" subcommand, TF_CLI_ARGS_apply should be ignored (apply != plan).
		testArgs(t, []string{"ltf", "plan", "-var", "one=1"}, []string{"TF_CLI_ARGS_apply=-var=two=2"}, arguments{
			bin:        "ltf",
			cli:        []string{"ltf", "plan", "-var", "one=1"},
			virtual:    []string{"ltf", "plan", "-var=one=1"},
			subcommand: "plan",
		})
	})
}

func TestArgumentsVersion(t *testing.T) {
	// Test with the flag, the correct way.
	testArgs(t, []string{"ltf", "-version"}, []string{}, arguments{
		bin:     "ltf",
		cli:     []string{"ltf", "-version"},
		virtual: []string{"ltf", "-version"},
		version: true,
	})

	// Test with a subcommand, also the correct way.
	testArgs(t, []string{"ltf", "version"}, []string{}, arguments{
		bin:        "ltf",
		cli:        []string{"ltf", "version"},
		virtual:    []string{"ltf", "version"},
		subcommand: "version",
		version:    true,
	})
}
