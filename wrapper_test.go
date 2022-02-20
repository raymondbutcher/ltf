package main

import (
	"fmt"
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
	Name    string         `hcl:"name,label"`
	Files   []string       `hcl:"files,optional"`
	Acts    []ActConfig    `hcl:"act,block"`
	Asserts []AssertConfig `hcl:"assert,block"`
}

type ActConfig struct {
	Name    string            `hcl:"name,label"`
	Env     map[string]string `hcl:"env,optional"`
	Cwd     string            `hcl:"cwd,optional"`
	Cmd     string            `hcl:"cmd"`
	Asserts []AssertConfig    `hcl:"assert,block"`
}

type AssertConfig struct {
	Name string            `hcl:"name,label"`
	Env  map[string]string `hcl:"env,optional"`
	Cmd  string            `hcl:"cmd,optional"`
}

func TestSuite(t *testing.T) {
	var tests TestConfig
	if err := hclsimple.DecodeFile("tests.hcl", nil, &tests); err != nil {
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

	for _, fileName := range arrange.Files {
		filePath := path.Join(tempDir, fileName)
		fileDir := path.Dir(filePath)
		err := os.MkdirAll(fileDir, os.ModePerm)
		is.NoErr(err) // error creating dir
		_, err = os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
		is.NoErr(err) // error creating file
	}

	// Act

	cwd := path.Join(tempDir, act.Cwd)
	args := strings.Split(act.Cmd, " ")
	env := []string{}
	for key, val := range act.Env {
		env = append(env, key+"="+val)
	}
	cmd, err := command(cwd, args, env, &Config{})
	is.NoErr(err)

	// Assert

	if assert.Cmd != "" {
		is.Equal(strings.Join(cmd.Args, " "), assert.Cmd) // ltf did not generate the expected command
	}

	if len(assert.Env) > 0 {
		for key, expected := range assert.Env {
			t.Run(key, func(t *testing.T) {
				is := is.New(t)
				actual := ""
				prefix := key + "="
				for _, env := range cmd.Env {
					if strings.HasPrefix(env, prefix) {
						actual = env[len(prefix):]
					}
				}
				is.Equal(actual, expected) // ltf did not set the expected environment variable
			})
		}
	}
}
