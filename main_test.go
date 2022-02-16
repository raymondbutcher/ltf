package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/matryer/is"
)

type TestConfig struct {
	Arranges []ArrangeConfig `hcl:"arrange,block"`
}

type ArrangeConfig struct {
	Files   []string       `hcl:"files,optional"`
	Acts    []ActConfig    `hcl:"act,block"`
	Asserts []AssertConfig `hcl:"assert,block"`
}

type ActConfig struct {
	Env     map[string]string `hcl:"env,optional"`
	Cwd     string            `hcl:"cwd,optional"`
	Cmd     string            `hcl:"cmd"`
	Asserts []AssertConfig    `hcl:"assert,block"`
}

type AssertConfig struct {
	Env map[string]string `hcl:"env,optional"`
	Cmd string            `hcl:"cmd,optional"`
}

func TestSuite(t *testing.T) {
	var tests TestConfig
	if err := hclsimple.DecodeFile("tests.hcl", nil, &tests); err != nil {
		log.Fatalf("Failed to load test suite: %s", err)
	}
	for i, arrange := range tests.Arranges {
		for j, act := range arrange.Acts {
			for k, assert := range act.Asserts {
				t.Run(fmt.Sprintf("arrange%d_act%d_nested_assert%d", i, j, k), func(t *testing.T) {
					runTestCase(t, arrange, act, assert)
				})
			}
			for k, assert := range arrange.Asserts {
				t.Run(fmt.Sprintf("arrange%d_act%d_top_assert%d", i, j, k), func(t *testing.T) {
					runTestCase(t, arrange, act, assert)
				})
			}
		}
	}
}

func runTestCase(t *testing.T, arrange ArrangeConfig, act ActConfig, assert AssertConfig) error {
	is := is.New(t)

	// Arrange

	for _, name := range arrange.Files {
		// TODO: create temp files for the test that get cleaned up after
		log.Printf("Create %s", name)
	}

	// Act

	// TODO: call real function
	cmd := exec.Cmd{}
	for key, val := range act.Env {
		log.Printf("Set %s=%s", key, val)
	}
	log.Printf("Set cwd to %s", act.Cwd)
	log.Printf("Run %s", act.Cmd)

	// Assert

	is.Equal(strings.Join(cmd.Args, " "), assert.Cmd) // ltf did not generate the exected command

	for key, expected := range assert.Env {
		t.Run(key, func(t *testing.T) {
			is := is.New(t)
			actual := ""
			prefix := fmt.Sprintf("%s=", key)
			for _, env := range cmd.Env {
				if strings.HasPrefix(env, prefix) {
					actual = env[len(prefix):]
				}
			}
			is.Equal(actual, expected) // ltf did not set the expected environment variable
		})
	}

	return nil
}

func TestHasConfFile(t *testing.T) {
	is := is.New(t)

	is.Equal(hasConfFile([]string{"one", "two", "three"}), false)
	is.Equal(hasConfFile([]string{"one", "two", "three.tf"}), true)
	is.Equal(hasConfFile([]string{"one", "two", "three.tf.json"}), true)
}
