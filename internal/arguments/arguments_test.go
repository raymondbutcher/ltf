package arguments

import (
	"testing"

	"github.com/matryer/is"
	"github.com/raymondbutcher/ltf"
)

func testArgs(t *testing.T, args []string, env ltf.Environ, expected Arguments) {
	is := is.New(t)

	// Already arranged

	// Act

	got, err := New(args, env)
	is.NoErr(err)

	// Assert

	is.Equal(got.Bin, expected.Bin)
	is.Equal(got.Args, expected.Args)
	is.Equal(got.Virtual, expected.Virtual)
	is.Equal(got.Chdir, expected.Chdir)
	is.Equal(got.Subcommand, expected.Subcommand)
	is.Equal(got.Help, expected.Help)
	is.Equal(got.Version, expected.Version)
}

func TestArgumentsChdir(t *testing.T) {
	testArgs(t, []string{"ltf", "-chdir=..", "plan"}, ltf.NewEnviron(), Arguments{
		Bin:        "ltf",
		Args:       []string{"ltf", "-chdir=..", "plan"},
		Virtual:    []string{"ltf", "-chdir=..", "plan"},
		Chdir:      "..",
		Subcommand: "plan",
	})
}

func TestArgumentsHelp(t *testing.T) {
	t.Run("flag", func(t *testing.T) {
		testArgs(t, []string{"ltf", "-help"}, ltf.NewEnviron(), Arguments{
			Bin:     "ltf",
			Args:    []string{"ltf", "-help"},
			Virtual: []string{"ltf", "-help"},
			Help:    true, // the flag should work
		})
	})

	t.Run("subcommand", func(t *testing.T) {
		testArgs(t, []string{"ltf", "help"}, ltf.NewEnviron(), Arguments{
			Bin:        "ltf",
			Args:       []string{"ltf", "help"},
			Virtual:    []string{"ltf", "help"},
			Subcommand: "help",
			Help:       false, // the subcommand is not correct usage
		})
	})
}

func TestArgumentsEmpty(t *testing.T) {
	testArgs(t, []string{"ltf"}, ltf.NewEnviron(), Arguments{
		Bin:     "ltf",
		Args:    []string{"ltf"},
		Virtual: []string{"ltf"},
	})
}

func TestArgumentsVars(t *testing.T) {
	t.Run("combined var arg", func(t *testing.T) {
		testArgs(t, []string{"ltf", "plan", "-var=one=1"}, ltf.NewEnviron(), Arguments{
			Bin:        "ltf",
			Args:       []string{"ltf", "plan", "-var=one=1"},
			Virtual:    []string{"ltf", "plan", "-var=one=1"},
			Subcommand: "plan",
		})
	})

	t.Run("separate var args", func(t *testing.T) {
		testArgs(t, []string{"ltf", "plan", "-var", "one=1"}, ltf.NewEnviron(), Arguments{
			Bin:        "ltf",
			Args:       []string{"ltf", "plan", "-var", "one=1"},
			Virtual:    []string{"ltf", "plan", "-var=one=1"},
			Subcommand: "plan",
		})
	})

	t.Run("combined var-file arg", func(t *testing.T) {
		testArgs(t, []string{"ltf", "plan", "-var-file=test.tfvars"}, ltf.NewEnviron(), Arguments{
			Bin:        "ltf",
			Args:       []string{"ltf", "plan", "-var-file=test.tfvars"},
			Virtual:    []string{"ltf", "plan", "-var-file=test.tfvars"},
			Subcommand: "plan",
		})
	})

	t.Run("separate var-file args", func(t *testing.T) {
		testArgs(t, []string{"ltf", "plan", "-var-file", "test.tfvars"}, ltf.NewEnviron(), Arguments{
			Bin:        "ltf",
			Args:       []string{"ltf", "plan", "-var-file", "test.tfvars"},
			Virtual:    []string{"ltf", "plan", "-var-file=test.tfvars"},
			Subcommand: "plan",
		})
	})

	// From the Terraform docs:
	// These arguments are inserted directly after the subcommand (such as plan)
	// and before any flags specified directly on the command-line. This behavior
	// ensures that flags on the command-line take precedence over environment variables.

	t.Run("env args", func(t *testing.T) {
		testArgs(t, []string{"ltf", "plan", "-var", "one=1"}, ltf.NewEnviron("TF_CLI_ARGS=-var=two=2 -var three=3"), Arguments{
			Bin:        "ltf",
			Args:       []string{"ltf", "plan", "-var", "one=1"},
			Virtual:    []string{"ltf", "plan", "-var=two=2", "-var=three=3", "-var=one=1"},
			Subcommand: "plan",
		})
	})

	t.Run("subcommand env args", func(t *testing.T) {
		testArgs(t, []string{"ltf", "plan", "-var", "one=1"}, ltf.NewEnviron("TF_CLI_ARGS_plan=-var=two=2"), Arguments{
			Bin:        "ltf",
			Args:       []string{"ltf", "plan", "-var", "one=1"},
			Virtual:    []string{"ltf", "plan", "-var=two=2", "-var=one=1"},
			Subcommand: "plan",
		})
	})

	t.Run("wrong subcommand env args", func(t *testing.T) {
		// When doing a "plan" subcommand, TF_CLI_ARGS_apply should be ignored (apply != plan).
		testArgs(t, []string{"ltf", "plan", "-var", "one=1"}, ltf.NewEnviron("TF_CLI_ARGS_apply=-var=two=2"), Arguments{
			Bin:        "ltf",
			Args:       []string{"ltf", "plan", "-var", "one=1"},
			Virtual:    []string{"ltf", "plan", "-var=one=1"},
			Subcommand: "plan",
		})
	})
}

func TestArgumentsVersion(t *testing.T) {
	// Test with the flag, the correct way.
	testArgs(t, []string{"ltf", "-version"}, ltf.NewEnviron(), Arguments{
		Bin:     "ltf",
		Args:    []string{"ltf", "-version"},
		Virtual: []string{"ltf", "-version"},
		Version: true,
	})

	// Test with a subcommand, also the correct way.
	testArgs(t, []string{"ltf", "version"}, ltf.NewEnviron(), Arguments{
		Bin:        "ltf",
		Args:       []string{"ltf", "version"},
		Virtual:    []string{"ltf", "version"},
		Subcommand: "version",
		Version:    true,
	})
}
