# LTF

> Status: alpha

LTF is a minimal, transparent Terraform wrapper. It makes Terraform projects easier to work with.

In standard Terraform projects, the `*.tf` files are typically duplicated in each environment, with only minor differences for the backend configuration. Every environment directory contains `*.tf` files and `*.tfvars` files. It requires some effort to maintain all of these environments and keep them consistent. Changes take longer because they involve more files.

```
terraform
├── dev
│   ├── dev.auto.tfvars
│   ├── main.tf
│   ├── outputs.tf
│   └── variables.tf
├── qa
│   ├── qa.auto.tfvars
│   ├── main.tf
│   ├── outputs.tf
│   └── variables.tf
└── live
    ├── blue
    │   ├── live.blue.auto.tfvars
    │   ├── main.tf
    │   ├── outputs.tf
    │   └── variables.tf
    └── green
        ├── live.green.auto.tfvars
        ├── main.tf
        ├── outputs.tf
        └── variables.tf
```

Using LTF, the `*.tf` files are shared between all environments. Environment directories only contain what is unique about the environment: the `*.tfvars` and `*.tfbackend` files. Maintenance is easier and environments are consistent by default. It takes less time to make changes because fewer files are involved.

```
ltf
├── main.tf
├── outputs.tf
├── variables.tf
├── dev
│   ├── dev.auto.tfvars
│   └── dev.tfbackend
├── qa
│   ├── qa.auto.tfvars
│   └── qa.tfbackend
└── live
    ├── blue
    │   ├── live.blue.auto.tfvars
    │   └── live.blue.tfbackend
    └── green
        ├── live.green.auto.tfvars
        └── live.green.tfbackend
```

Using LTF is very easy. It avoids tedious command line arguments. Change to an environment directory and use `ltf` just like you would normally use `terraform` in a simple, single-directory Terraform project.

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

## Why use LTF?

LTF only does a few things:

* It finds and uses a parent directory as the configuration directory.
* It finds and uses tfvars and tfbackend files from the current and parent directories.
* It supports hooks to run custom scripts before and after Terraform.

LTF is good because:

* It avoids tedious command line arguments, so it's quick and easy to use.
* It is a transparent wrapper, so all Terraform actions and arguments can be used.
* It is released as a single binary, so installation is easy.
* It keeps your Terraform configuration DRY using only a simple directory structure.
* It only does a few things, so there's not much to learn.
* It runs Terraform in the configuration directory, so there's no extra build/cache directory to complicate things.

LTF is purposefully simple and feature-light, so it doesn't do everything:

* [It does not create backend resources for you.](https://github.com/raymondbutcher/ltf/issues/11)
* [It does not generate Terraform configuration using another language.](https://github.com/raymondbutcher/ltf/issues/12)
* [It does not support module/stack/state dependencies or run-all commands.](https://github.com/raymondbutcher/ltf/issues/13)
* [It does not support remote configurations.](https://github.com/raymondbutcher/ltf/issues/14)

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
* Sets the `TF_DATA_DIR` environment variable to make Terraform use the `.terraform` directory inside the current directory, rather than in the configuration directory.

It always does the following:

* Finds `*.tfvars` and `*.tfvars.json` files in the current directory and parent directories, stopping at the configuration directory, then sets the `TF_VAR_name` environment variable for each variable it finds.
  * Terraform's [precedence rules](https://www.terraform.io/language/values/variables#variable-definition-precedence) are followed when finding variables. Variables in subdirectories will take precendence over variables in parent directories.
  * If any tfvars files exist in the configuration directory, Terraform will use those values instead of the environment variables set by LTF. LTF raises an error if the environment variable does not match the value that Terraform will use. This can be avoided by using variable defaults instead of tfvars files, or by moving the tfvars files into a subdirectory.
* Finds `*.tfbackend` files in the current directory and parent directories, stopping at the configuration directory, then updates the `TF_CLI_ARGS_init` environment variable to contain `-backend-config=$file` for each file.
* Runs hook scripts before and after Terraform.

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
