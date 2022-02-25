package main

// version is updated with -ldflags in the pipeline.
var version = "development"

func getVersion() string {
	return version
}
