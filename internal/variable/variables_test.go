package variable

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/matryer/is"
	"github.com/raymondbutcher/ltf/internal/arguments"
	"github.com/zclconf/go-cty/cty"
)

func TestLoad(t *testing.T) {
	is := is.New(t)

	// Arrange

	tempDir, err := os.MkdirTemp("", "ltf-test-")
	is.NoErr(err) // error creating temporary directory
	defer os.RemoveAll(tempDir)

	contents, err := ioutil.ReadFile("variables_test.tf")
	is.NoErr(err) // error reading file
	err = ioutil.WriteFile(path.Join(tempDir, "main.tf"), contents, 06666)
	is.NoErr(err) // error creating file

	contents, err = ioutil.ReadFile("variables_test.tfvars")
	is.NoErr(err) // error reading file
	err = ioutil.WriteFile(path.Join(tempDir, "terraform.tfvars"), contents, 06666)
	is.NoErr(err) // error creating file

	args, err := arguments.New([]string{"ltf"}, []string{})
	is.NoErr(err) // error creating arguments

	// Act

	vars, err := Load(args, []string{tempDir}, tempDir)
	is.NoErr(err) // error loading variables

	// Assert

	is.Equal(vars["bool_default_value"].StringValue, "true")
	is.Equal(vars["bool_default_value"].AnyValue, cty.BoolVal(true))

	is.Equal(vars["bool_no_value"].StringValue, "")
	is.Equal(vars["bool_no_value"].AnyValue, cty.NilVal)

	is.Equal(vars["bool_value"].StringValue, "true")
	is.Equal(vars["bool_value"].AnyValue, cty.BoolVal(true))

	is.Equal(vars["list_default_value"].StringValue, "[true]")
	is.Equal(vars["list_default_value"].AnyValue, cty.TupleVal([]cty.Value{cty.BoolVal(true)}))

	is.Equal(vars["list_no_value"].StringValue, "")
	is.Equal(vars["list_no_value"].AnyValue, cty.NilVal)

	is.Equal(vars["list_value"].StringValue, "[true,true]")
	is.Equal(vars["list_value"].AnyValue, cty.TupleVal([]cty.Value{cty.BoolVal(true), cty.BoolVal(true)}))

	is.Equal(vars["string_default_value"].StringValue, "string_default_value")
	is.Equal(vars["string_default_value"].AnyValue, cty.StringVal("string_default_value"))

	is.Equal(vars["string_no_value"].StringValue, "")
	is.Equal(vars["string_no_value"].AnyValue, cty.StringVal(""))

	is.Equal(vars["string_value"].StringValue, "string_value")
	is.Equal(vars["string_value"].AnyValue, cty.StringVal("string_value"))

	is.Equal(vars["untyped_no_value"].StringValue, "")
	is.Equal(vars["untyped_no_value"].AnyValue, cty.StringVal(""))

	is.Equal(vars["untyped_default_list_value"].StringValue, `["untyped_default_list_value"]`)
	is.Equal(vars["untyped_default_list_value"].AnyValue, cty.StringVal(`["untyped_default_list_value"]`))

	is.Equal(vars["untyped_bool_value"].StringValue, "true")
	is.Equal(vars["untyped_bool_value"].AnyValue, cty.StringVal("true"))

	is.Equal(vars["untyped_string_value"].StringValue, "untyped_string_value")
	is.Equal(vars["untyped_string_value"].AnyValue, cty.StringVal("untyped_string_value"))
}
