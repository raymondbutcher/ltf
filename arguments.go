package ltf

// Arguments contains information about the arguments passed into the LTF
// program via command line arguments and the TF_CLI_ARGS
// and TF_CLI_ARGS_name environment variables.
type Arguments struct {
	// Bin is the command that was run.
	Bin string

	// Args holds command line arguments, including the value of Bin as Args[0].
	Args []string

	// Virtual holds the combined arguments from Args and also extra arguments
	// provided by the `TF_CLI_ARGS` and `TF_CLI_ARGS_name` environment variables.
	Virtual []string

	// EnvToJson is true if the -env-to-json flag was specified.
	// This is a special LTF flag used by the hooks system.
	EnvToJson bool

	// Chdir is the value of the -chdir global option if specified.
	Chdir string

	// Help is true if the -help global option is specified.
	Help bool

	// Version is true if the -version global option is specified
	// or the subcommand is "version".
	Version bool

	// Subcommand is the first non-flag argument if there is one.
	Subcommand string
}
