package ltf

import (
	"testing"

	"github.com/matryer/is"
)

func TestEnvFunctions(t *testing.T) {
	is := is.New(t)
	env := NewEnviron([]string{
		"ONE=1",
		"TWO=2",
		"THREE=3",
	})

	is.Equal(env.GetValue("ONE"), "1")
	is.Equal(env.GetValue("TWO"), "2")
	is.Equal(env.GetValue("THREE"), "3")

	env.SetValue("TWO", "changed")

	is.Equal(env.GetValue("ONE"), "1")
	is.Equal(env.GetValue("TWO"), "changed")
	is.Equal(env.GetValue("THREE"), "3")

	is.Equal(len(env.env), 3)
}
