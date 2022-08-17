package terraform

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/matryer/is"
	"github.com/zclconf/go-cty/cty"
)

func TestLoad(t *testing.T) {
	is := is.New(t)

	// Arrange

	tempDir, err := os.MkdirTemp("", "ltf-test-")
	is.NoErr(err) // error creating temporary directory
	defer os.RemoveAll(tempDir)

	contents, err := ioutil.ReadFile("variable_test.tf")
	is.NoErr(err) // error reading file
	err = ioutil.WriteFile(path.Join(tempDir, "main.tf"), contents, 06666)
	is.NoErr(err) // error creating file

	contents, err = ioutil.ReadFile("variable_test.tfvars")
	is.NoErr(err) // error reading file
	err = ioutil.WriteFile(path.Join(tempDir, "terraform.tfvars"), contents, 06666)
	is.NoErr(err) // error creating file

	args, err := NewArguments([]string{"ltf"}, []string{})
	is.NoErr(err) // error creating arguments

	// Act

	vars := NewVariableService()
	err = vars.Load(args, []string{tempDir}, tempDir)
	is.NoErr(err) // error loading variables

	// Assert

	is.Equal(vars.GetValue("bool_default_value"), "true")
	is.Equal(vars.GetCtyValue("bool_default_value"), cty.BoolVal(true))

	is.Equal(vars.GetValue("bool_no_value"), "")
	is.Equal(vars.GetCtyValue("bool_no_value"), cty.NilVal)

	is.Equal(vars.GetValue("bool_value"), "true")
	is.Equal(vars.GetCtyValue("bool_value"), cty.BoolVal(true))

	is.Equal(vars.GetValue("list_default_value"), "[true]")
	is.Equal(vars.GetCtyValue("list_default_value"), cty.TupleVal([]cty.Value{cty.BoolVal(true)}))

	is.Equal(vars.GetValue("list_no_value"), "")
	is.Equal(vars.GetCtyValue("list_no_value"), cty.NilVal)

	is.Equal(vars.GetValue("list_value"), "[true,true]")
	is.Equal(vars.GetCtyValue("list_value"), cty.TupleVal([]cty.Value{cty.BoolVal(true), cty.BoolVal(true)}))

	is.Equal(vars.GetValue("string_default_value"), "string_default_value")
	is.Equal(vars.GetCtyValue("string_default_value"), cty.StringVal("string_default_value"))

	is.Equal(vars.GetValue("string_no_value"), "")
	is.Equal(vars.GetCtyValue("string_no_value"), cty.StringVal(""))

	is.Equal(vars.GetValue("string_value"), "string_value")
	is.Equal(vars.GetCtyValue("string_value"), cty.StringVal("string_value"))

	is.Equal(vars.GetValue("untyped_no_value"), "")
	is.Equal(vars.GetCtyValue("untyped_no_value"), cty.StringVal(""))

	is.Equal(vars.GetValue("untyped_default_list_value"), `["untyped_default_list_value"]`)
	is.Equal(vars.GetCtyValue("untyped_default_list_value"), cty.StringVal(`["untyped_default_list_value"]`))

	is.Equal(vars.GetValue("untyped_bool_value"), "true")
	is.Equal(vars.GetCtyValue("untyped_bool_value"), cty.StringVal("true"))

	is.Equal(vars.GetValue("untyped_string_value"), "untyped_string_value")
	is.Equal(vars.GetCtyValue("untyped_string_value"), cty.StringVal("untyped_string_value"))
}
