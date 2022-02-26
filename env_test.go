package main

import (
	"testing"

	"github.com/matryer/is"
)

func TestEnvFunctions(t *testing.T) {
	is := is.New(t)
	env := []string{
		"ONE=1",
		"TWO=2",
		"THREE=3",
	}
	env = setEnvValue(env, "TWO", "changed")
	is.Equal(getEnvValue(env, "TWO"), "changed")
}
