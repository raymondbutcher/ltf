package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/matryer/is"
)

type TestConfig struct {
	Arranges []ArrangeConfig `hcl:"arrange,block"`
}

type ArrangeConfig struct {
	Name    string            `hcl:"name,label"`
	Files   map[string]string `hcl:"files,optional"`
	Acts    []ActConfig       `hcl:"act,block"`
	Asserts []AssertConfig    `hcl:"assert,block"`
}

type ActConfig struct {
	Name    string            `hcl:"name,label"`
	Env     map[string]string `hcl:"env,optional"`
	Cwd     string            `hcl:"cwd,optional"`
	Cmd     string            `hcl:"cmd"`
	Asserts []AssertConfig    `hcl:"assert,block"`
}

type AssertConfig struct {
	Name     string            `hcl:"name,label"`
	Cmd      string            `hcl:"cmd,optional"`
	Env      map[string]string `hcl:"env,optional"`
	ExitCode int               `hcl:"exit,optional"`
}

func TestSuite(t *testing.T) {
	var tests TestConfig
	if err := hclsimple.DecodeFile("ltf_test.hcl", nil, &tests); err != nil {
		log.Fatalf("Failed to load test suite: %s", err)
	}
	for _, arrange := range tests.Arranges {
		for _, act := range arrange.Acts {
			for _, assert := range append(act.Asserts, arrange.Asserts...) {
				name := fmt.Sprintf("arrange.%s/act.%s/assert.%s", arrange.Name, act.Name, assert.Name)
				t.Run(name, func(t *testing.T) {
					runTestCase(t, arrange, act, assert)
				})
			}
		}
	}
}

func runTestCase(t *testing.T, arrange ArrangeConfig, act ActConfig, assert AssertConfig) {
	is := is.New(t)

	// Arrange

	tempDir, err := os.MkdirTemp("", "ltf-test-")
	is.NoErr(err)
	defer os.RemoveAll(tempDir)

	for fileName, fileContents := range arrange.Files {
		filePath := path.Join(tempDir, fileName)
		fileDir := path.Dir(filePath)
		err := os.MkdirAll(fileDir, os.ModePerm)
		is.NoErr(err) // error creating dir
		err = ioutil.WriteFile(filePath, []byte(fileContents), 06666)
		is.NoErr(err) // error creating file
	}

	// Act

	cwd := path.Join(tempDir, act.Cwd)
	env := []string{"LTF_TEST_MODE=1"}
	args, err := newArguments(strings.Split(act.Cmd, " "), env)
	is.NoErr(err) // error parsing arguments
	for key, val := range act.Env {
		env = append(env, key+"="+val)
	}
	cmd, exitCode, err := ltf(cwd, args, env)
	is.NoErr(err)

	// Assert

	is.Equal(exitCode, assert.ExitCode) // ltf exited with unexpected code

	if assert.Cmd != "" {
		is.Equal(strings.Join(cmd.Args, " "), assert.Cmd) // ltf did not generate the expected command
	}

	if len(assert.Env) > 0 {
		for name, expected := range assert.Env {
			t.Run(name, func(t *testing.T) {
				is := is.New(t)
				actual := getEnvValue(cmd.Env, name)
				is.Equal(actual, expected) // ltf did not set the expected environment variable
			})
		}
	}
}
