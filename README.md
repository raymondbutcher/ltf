# LTF

> Status: alpha

LTF is a lightweight, transparent Terraform wrapper. It makes Terraform projects easier to work with, and makes the Terraform command easier to use.

Features:

* All standard Terraform command line options can be used.
* Finds and uses a parent directory as the configuration directory.
* Finds and uses tfvars and tfbackend files from the current and parent directories.
* Hooks to run commands before/after Terraform.

A standard LTF project might look like this:

```
example <----------------------------- configuration directory
├── ltf.yaml                           LTF configuration (optional, for hooks)
├── main.tf                            configuration file(s)
├── dev
│   ├── dev.auto.tfvars
│   └── dev.tfbackend
└── live <---------------------------- intermediate directory
    ├── live.auto.tfvars               variables file(s)
    ├── live.tfbackend                 backend file(s)
    ├── blue
    │   ├── live.blue.auto.tfvars
    │   └── live.blue.tfbackend
    └── green <----------------------- working directory
        ├── live.green.auto.tfvars     variables file(s)
        └── live.green.tfbackend       backend file(s)
```

Typical usage would look like this:

```
$ cd dev
$ ltf init
$ ltf plan
$ ltf apply
$ cd ../live/green
$ ltf init
$ ltf plan
$ ltf apply -target=random_id.this
```

## Why choose LTF over other approaches?

LTF has these benefits:

* LTF is a transparent wrapper, so all Terraform actions and arguments can be used.
* LTF is released as a single binary, so installation is easy.
* LTF keeps your configuration DRY using only the directory structure.
* LTF requires almost no learning to use.
* LTF runs Terraform in the current working directory, so there's no build/cache directory to complicate things.

But LTF does not aim to do everything:

* [LTF does not create backend resources for you.](https://github.com/raymondbutcher/ltf/issues/11)
* [LTF does not generate Terraform configuration using another language.](https://github.com/raymondbutcher/ltf/issues/12)
* [LTF does not support module/stack/state dependencies or run-all commands.](https://github.com/raymondbutcher/ltf/issues/13)
* [LTF does not support remote configurations.](https://github.com/raymondbutcher/ltf/issues/14)

## Installation

### Install manually

Download the appropriate binary for your system from the [releases](https://github.com/raymondbutcher/ltf/releases) page, move it to your `PATH`, and make it executable.

### Install using [bin](https://github.com/marcosnils/bin)

[bin](https://github.com/marcosnils/bin) manages binary files downloaded from different sources. Run the following to install the latest version of LTF:

```
bin install github.com/raymondbutcher/ltf
```

### Verify installation

Run the following to verify that LTF has been installed:

```
ltf -help
```

## Usage

Run `ltf` instead of `terraform`.

Example:

```
$ ltf init
$ ltf plan
$ ltf apply -target=random_id.this
```

## How it works

LTF searches the directory tree for a Terraform configuration directory, tfvars files, and tfbackend files, and passes those details to Terraform.

When LTF finds no `*.tf` or `*.tf.json` files in the current directory, it does the following:

* Finds the closest parent directory containing `*.tf` or `*.tf.json` files, then adds `-chdir=$dir` to the Terraform command line arguments, to make Terraform use that directory as the configuration directory.
* Updates the `TF_DATA_DIR` environment variable to make Terraform use the `.terraform` directory inside the current directory, rather than in the configuration directory.

It also does the following:

* Finds `*.tfvars` and `*.tfvars.json` files in the current directory and parent directories, stopping at the configuration directory, then updates the `TF_CLI_ARGS_plan` and `TF_CLI_ARGS_apply` environment variables to contain `-var-file=$file` for each file.
  * Terraform's [precedence rules](https://www.terraform.io/language/values/variables#variable-definition-precedence) are followed when finding files to use, except that files in subdirectories will take precendence over files in parent directories.
* Finds `*.tfbackend` files in the current directory and parent directories, stopping at the configuration directory, then updates the `TF_CLI_ARGS_init` environment variable to contain `-backend-config=$file` for each file.

## Hooks

LTF also supports hook scripts defined in `ltf.yaml`. It looks for this file in the current directory or any parent directory. Hook scripts are just Bash scripts; they can contain multiple lines, and they can even export environment variables. Environment variables will persist to subsequent hooks and the Terraform command.

Hooks can be configured to run `before` specific Terraform commands, and/or `after` they have completed successfully, and/or after they have `failed`.

### Schema

```yaml
hooks:
  $name: # the name of the hook
    before: # (optional) run the script before these commands:
      - terraform # the hook will always run
      - terraform $subcommand # the hook will only run before this subcommand
    after: [] # (optional) run the script after these commands finish successfully
    failed: [] # (optional) run the script after these commands have failed
    script: $script # bash script to run
```

### Example: running commands

```yaml
hooks:
  greetings:
    before:
      - terraform
    script: |
      echo "Hello, this is a hook!"
      echo "The date is $(date -I)"
```

### Example: Setting environment variables

```yaml
hooks:
  TF_VAR_hook:
    before:
      - terraform apply
      - terraform plan
    script: export TF_VAR_hook=hello
```
