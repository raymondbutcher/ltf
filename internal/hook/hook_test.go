package hook

import (
	"testing"

	"github.com/matryer/is"
	"github.com/raymondbutcher/ltf"
	"github.com/raymondbutcher/ltf/internal/arguments"
)

func TestHookMatch(t *testing.T) {
	is := is.New(t)

	t.Run("only before terraform", func(t *testing.T) {
		// Arrange

		h := Hook{
			Before: []string{"terraform"},
			After:  []string{},
			Failed: []string{},
		}

		t.Run("with subcommand", func(t *testing.T) {
			// Act

			args, err := arguments.New([]string{"terraform", "plan"}, ltf.NewEnviron())
			is.NoErr(err)

			// Assert

			is.Equal(h.Match("before", args), true)
			is.Equal(h.Match("after", args), false)
			is.Equal(h.Match("failed", args), false)
		})

		t.Run("without subcommand", func(t *testing.T) {
			// Act

			args, err := arguments.New([]string{"terraform"}, ltf.NewEnviron())
			is.NoErr(err)

			// Assert

			is.Equal(h.Match("before", args), true)
			is.Equal(h.Match("after", args), false)
			is.Equal(h.Match("failed", args), false)
		})

	})

	t.Run("only after terraform apply", func(t *testing.T) {
		// Arrange

		h := Hook{
			Before: []string{"terraform init"},
			After:  []string{"terraform init", "terraform apply"},
			Failed: []string{"terraform init"},
		}

		t.Run("with subcommand", func(t *testing.T) {
			// Act

			args, err := arguments.New([]string{"terraform", "apply"}, ltf.NewEnviron())
			is.NoErr(err)

			// Assert

			is.Equal(h.Match("before", args), false)
			is.Equal(h.Match("after", args), true)
			is.Equal(h.Match("failed", args), false)
		})

		t.Run("without subcommand", func(t *testing.T) {
			// Act

			args, err := arguments.New([]string{"terraform"}, ltf.NewEnviron())
			is.NoErr(err)

			// Assert

			is.Equal(h.Match("before", args), false)
			is.Equal(h.Match("after", args), false)
			is.Equal(h.Match("failed", args), false)
		})

		t.Run("with wrong subcommand", func(t *testing.T) {
			// Act

			args, err := arguments.New([]string{"terraform", "plan"}, ltf.NewEnviron())
			is.NoErr(err)

			// Assert

			is.Equal(h.Match("before", args), false)
			is.Equal(h.Match("after", args), false)
			is.Equal(h.Match("failed", args), false)
		})
	})
}
