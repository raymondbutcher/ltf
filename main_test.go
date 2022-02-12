package main

import (
	"testing"

	"github.com/matryer/is"
)

func TestRequestHandler(t *testing.T) {
	is := is.New(t)

	is.Equal("hello", "hello")
}
