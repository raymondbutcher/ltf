# LTF

> Status: partially implemented, untested

LTF is a lightweight Terraform wrapper that adds minimal features to make it easier to work with Terraform projects.

Features:

* Automatically use configuration files from a parent directory.
* Automatically find and use tfbackend and tfvars files.

LTF is a *transparent* wrapper, so all standard Terraform command line options can be used when using LTF.

A standard LTF project looks like this:

```
.
├── ecr <--------------------------------------------- configuration directory
│   ├── main.tf <------------------------------------- configuration file(s)
│   ├── dev
│   │   ├── ecr.dev.auto.tfvars
│   │   └── ecr.dev.s3.tfbackend
│   └── live
│       ├── eu-central-1
│       │   ├── ecr.live.eu-central-1.auto.tfvars
│       │   └── ecr.live.eu-central-1.s3.tfbackend
│       └── eu-west-1 <------------------------------- working directory
│           ├── ecr.live.eu-west-1.auto.tfvars <------ variables file(s)
│           └── ecr.live.eu-west-1.s3.tfbackend <----- backend file(s)
└── iam
    ├── main.tf
    ├── dev
    │   ├── iam.dev.auto.tfvars
    │   └── iam.dev.s3.tfbackend
    └── live
        ├── iam.live.auto.tfvars
        └── iam.live.s3.tfbackend
```

Typical usage would look like this:

```
$ cd ecr/live/eu-central-1
$ ltf init
$ ltf plan
$ ltf apply
$ cd ../eu-west-1
$ ltf init
$ ltf plan
$ ltf apply -target=aws_ecr_repository.this
```

## Why choose LTF over other approaches?

LTF has these benefits:

* LTF is a transparent wrapper, so all Terraform actions and arguments can be used.
* LTF is released as a single binary, so installation is easy.
* LTF keeps your configuration DRY, using a simple project structure with no extra files.
* LTF requires minimal learning to use.
* LTF runs Terraform in the current working directory, so there's no build/cache directory to complicate things.

But LTF does not aim to do everything:

* LTF does not create backend resources for you (see Pretf, Terragrunt, Terraspace).
* LTF does not support generating Terraform configuration using another language (see Pretf, Terraspace).
* LTF does not support module/stack/state dependencies (see Tau, Terragrunt, Terraspace).
* LTF does not support remote configurations (see Pretf, Tau, Terragrunt).
* LTF does not support run-all or similar (see Tau, Terragrunt, Terraspace).

## Installation

LTF is released as a single binary. Download and add it to your `PATH`. Make it executable if needed.

## Usage

Run `ltf` instead of `terraform`.

Example:

```
$ ltf init
$ ltf plan
$ ltf apply -target=aws_ecr_repository.this
```

## How it works

LTF searches the directory tree for a Terraform configuration directory, tfvars files, and tfbackend files, and passes those details to Terraform.

When LTF finds no `*.tf` and `*.tf.json` files in the current directory, it does the following:

* Finds the first parent directory containing `*.tf` or `*.tf.json` files and adds `-chdir=$dir` to the Terraform command line arguments, to make Terraform change to that directory when it runs.

It also does the following:

* Updates the `TF_DATA_DIR` environment variable to make Terraform use the `.terraform` directory inside the current directory, rather than in the configuration directory.
* Finds `*.tfvars` and `*.tfvars.json` files in the current directory and parent directories, stopping at the configuration directory, then updates the `TF_CLI_ARGS_plan` and `TF_CLI_ARGS_apply` environment variables to contain `-var-file=$filename` for each file. Terraform's [rules](https://www.terraform.io/language/values/variables#variable-definition-precedence) are followed when finding files to use.
* Finds `*.tfbackend` files in the current directory and parent directories, stopping at the configuration directory, then updates the `TF_CLI_ARGS_init` environment variable to contain `-backend-config=$filename` for each file.
