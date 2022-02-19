# LTF

> Status: pre-alpha

LTF is a lightweight, transparent Terraform wrapper. It makes Terraform projects easier to work with, and makes the Terraform command easier to run.

Features:

* All standard Terraform command line options can be used.
* Finds and uses a parent directory as the configuration directory.
* Finds and uses tfvars and tfbackend files from the current and parent directories.

A standard LTF project might look like this:

```
example <------------------------------ configuration directory
├── main.tf                             configuration file(s)
├── dev
│   ├── dev.auto.tfvars
│   └── dev.tfbackend
└── live <----------------------------- intermediate directory
    ├── live.auto.tfvars                variables file(s)
    ├── blue
    │   ├── live.blue.auto.tfvars
    │   └── live.blue.tfbackend
    └── green <------------------------ working directory
        ├── live.green.auto.tfvars      variables file(s)
        └── live.green.tfbackend        backend file(s)
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

LTF leverages a simple projects structure and Terraform's 

This simple, logical project structure could be used with Terraform directly. However, it would require extra command line arguments and environment variables

 combined with LTF allows for DRY configuration while retaining full, standard Terraform command line usage. Less is more!

## Why choose LTF over other approaches?

LTF has these benefits:

* LTF is a transparent wrapper, so all Terraform actions and arguments can be used.
* LTF is released as a single binary, so installation is easy.
* LTF keeps your configuration DRY using only the directory structure.
* LTF requires almost no learning to use.
* LTF runs Terraform in the current working directory, so there's no build/cache directory to complicate things.

But LTF does not aim to do everything:

* LTF does not create backend resources for you (see Pretf, Terragrunt, Terraspace).
* LTF does not generate Terraform configuration using another language (see CDK, Pretf, Terraspace).
* LTF does not support apply-all/run-all/etc (see Tau, Terragrunt, Terraspace).
* LTF does not support module/stack/state dependencies (see Tau, Terragrunt, Terraspace).
* LTF does not support remote configurations (see Pretf, Tau, Terragrunt).

## Installation

> TODO: add releases and link to the releases page here

> TODO: try out https://github.com/marwanhawari/stew and https://github.com/marcosnils/bin and if good, add isntructions here

> TODO: consider creating asdf plugin and adding instructions here

LTF is released as a single binary. Download and add it to your `PATH`. Make it executable if needed.

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
