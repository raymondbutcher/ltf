package main

import (
	"testing"

	"github.com/matryer/is"
)

func TestHasConfFile(t *testing.T) {
	is := is.New(t)

	is.Equal(hasConfFile([]string{"one", "two", "three"}), false)
	is.Equal(hasConfFile([]string{"one", "two", "three.tf"}), true)
	is.Equal(hasConfFile([]string{"one", "two", "three.tf.json"}), true)
}

func TestHasVarsFile(t *testing.T) {
	is := is.New(t)

	is.Equal(hasVarsFile([]string{"one", "two", "three"}), false)
	is.Equal(hasVarsFile([]string{"one", "two", "three.tfvars"}), true)
	is.Equal(hasVarsFile([]string{"one", "two", "three.tfvars.json"}), true)
}
