package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/matryer/is"
)

const backendContents = `
bucket = "some-bucket"
key    = "${var.stack}/terraform.tfstate"
region = var.region
`

var backendVars = variables{
	"stack":  {name: "stack", value: "vpc"},
	"region": {name: "region", value: "eu-west-1"},
}

func TestParseBackendFile(t *testing.T) {
	is := is.New(t)

	// Arrange

	tempDir, err := os.MkdirTemp("", "ltf-test-")
	is.NoErr(err) // error creating temporary directory
	defer os.RemoveAll(tempDir)

	filename := path.Join(tempDir, "s3.tfbackend")
	err = ioutil.WriteFile(filename, []byte(backendContents), 06666)
	is.NoErr(err) // error creating file

	// Act

	values, err := parseBackendFile(filename, backendVars)
	is.NoErr(err)

	// Assert

	is.Equal(values["bucket"], "some-bucket")
	is.Equal(values["key"], "vpc/terraform.tfstate")
	is.Equal(values["region"], "eu-west-1")
}
