package backend

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/matryer/is"
	"github.com/raymondbutcher/ltf"
	"github.com/raymondbutcher/ltf/internal/terraform"
)

func TestParseBackendFile(t *testing.T) {
	is := is.New(t)

	// Arrange

	tempDir, err := os.MkdirTemp("", "ltf-test-")
	is.NoErr(err) // error making temporary directory
	defer os.RemoveAll(tempDir)

	contents, err := ioutil.ReadFile("backend_test.tfbackend")
	is.NoErr(err) // error reading file
	filename := path.Join(tempDir, "s3.tfbackend")
	err = ioutil.WriteFile(filename, contents, 06666)
	is.NoErr(err) // error writing file

	args, err := terraform.NewArguments([]string{"ltf"}, ltf.Environ{})
	is.NoErr(err) // error creating arguments
	vars := terraform.NewVariableService()
	err = vars.Load(args, []string{"."}, ".")
	is.NoErr(err) // error loading variables

	// Act

	values, err := parseBackendFile(filename, vars)
	is.NoErr(err)

	// Assert

	is.Equal(values["bucket"], "some-bucket")
	is.Equal(values["key"], "vpc/terraform.tfstate")
	is.Equal(values["region"], "eu-west-1")
	is.Equal(values["extra"], "success")
	is.Equal(values["encrypted"], "true")
}
