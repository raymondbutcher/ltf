package main

import (
	"path"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// findBackendFiles returns backend files to use in the Terraform command.
func findBackendFiles(dirs []string, chdir string) (backendFiles []string, err error) {

	// Start at the highest directory (configuration directory)
	// and go deeper towards the current directory.
	// Files in the current directory take precedence
	// over files in parent directories.
	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]

		// Get a sorted list of files in this directory.
		files, err := getFileNames(dir)
		if err != nil {
			return nil, err
		}
		sort.Strings(files)

		// Add any matching backend files.
		for _, name := range matchFiles(files, "*.tfbackend") {
			backendFiles = append(backendFiles, path.Join(dir, name))
		}
	}

	return backendFiles, nil
}

// parseBackendFile parses a *.tfbackend file as HCL into a map of strings.
// Variables can be used in the same way as *.tf files using the `var` object.
func parseBackendFile(filename string, vars map[string]string) (map[string]string, error) {
	// Parse the file.
	p := hclparse.NewParser()
	file, diags := p.ParseHCLFile(filename)
	for _, err := range diags.Errs() {
		return nil, err // TODO: handle multiple errors
	}

	// Decode the file into a map of cty values.
	values := map[string]cty.Value{}
	ctx := varsEvalContext(vars)
	diags = gohcl.DecodeBody(file.Body, ctx, &values)
	for _, err := range diags.Errs() {
		return nil, err // TODO: handle multiple errors
	}

	// Convert the cty values into strings.
	strings := map[string]string{}
	for key, val := range values {
		var s string
		err := gocty.FromCtyValue(val, &s)
		if err != nil {
			return nil, err
		}
		strings[key] = s
	}

	return strings, nil
}

// varsEvalContext returns an EvalContext with a `var` variable
// containing all of the variables.
func varsEvalContext(vars map[string]string) *hcl.EvalContext {
	ctx := hcl.EvalContext{}
	values := map[string]cty.Value{}
	for name, val := range vars {
		values[name] = cty.StringVal(val)
	}
	ctx.Variables = map[string]cty.Value{"var": cty.ObjectVal(values)}
	return &ctx
}
