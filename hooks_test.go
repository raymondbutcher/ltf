package main

import (
	"testing"

	"github.com/matryer/is"
)

func TestHookMatch(t *testing.T) {
	is := is.New(t)

	t.Run("only before terraform", func(t *testing.T) {
		// Arrange

		h := hook{
			Before: []string{"terraform"},
			After:  []string{},
			Failed: []string{},
		}

		t.Run("with subcommand", func(t *testing.T) {
			// Act

			args, err := newArguments([]string{"terraform", "plan"}, []string{})
			is.NoErr(err)

			// Assert

			is.Equal(h.match("before", args), true)
			is.Equal(h.match("after", args), false)
			is.Equal(h.match("failed", args), false)
		})

		t.Run("without subcommand", func(t *testing.T) {
			// Act

			args, err := newArguments([]string{"terraform"}, []string{})
			is.NoErr(err)

			// Assert

			is.Equal(h.match("before", args), true)
			is.Equal(h.match("after", args), false)
			is.Equal(h.match("failed", args), false)
		})

	})

	t.Run("only after terraform apply", func(t *testing.T) {
		// Arrange

		h := hook{
			Before: []string{"terraform init"},
			After:  []string{"terraform init", "terraform apply"},
			Failed: []string{"terraform init"},
		}

		t.Run("with subcommand", func(t *testing.T) {
			// Act

			args, err := newArguments([]string{"terraform", "apply"}, []string{})
			is.NoErr(err)

			// Assert

			is.Equal(h.match("before", args), false)
			is.Equal(h.match("after", args), true)
			is.Equal(h.match("failed", args), false)
		})

		t.Run("without subcommand", func(t *testing.T) {
			// Act

			args, err := newArguments([]string{"terraform"}, []string{})
			is.NoErr(err)

			// Assert

			is.Equal(h.match("before", args), false)
			is.Equal(h.match("after", args), false)
			is.Equal(h.match("failed", args), false)
		})

		t.Run("with wrong subcommand", func(t *testing.T) {
			// Act

			args, err := newArguments([]string{"terraform", "plan"}, []string{})
			is.NoErr(err)

			// Assert

			is.Equal(h.match("before", args), false)
			is.Equal(h.match("after", args), false)
			is.Equal(h.match("failed", args), false)
		})
	})
}
