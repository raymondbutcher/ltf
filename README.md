# LTF

> Status: largely not implemented and untested

LTF is a lightweight Terraform wrapper that adds minimal features to make it easier to work with Terraform projects.

Features:

* DRY configuration:
  * Use variables files from the current directory with configuration files from a parent directory.
* Dynamic backends:
  * Allow use of variables in backend configuration blocks.
  * Automatically use tfbackend files from the current directory.

LTF is a *transparent* wrapper, so all standard Terraform command line options can be used when using LTF. LTF does not get in the way.

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
* LTF does not support module/stack/state dependencies (see Terragrunt, Terraspace).
* LTF does not support remote configurations (see Pretf, Terragrunt).
* LTF does not support run-all or similar (see Terragrunt, Terraspace).

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

## Feature: DRY configuration

LTF makes it easy to work with multiple environments or deployments of the same configuration. It does this by using variables from the current directory with configuration from a parent directory.

<details>
  <summary>How does it work?</summary>

> When LTF runs and finds no `tf` files in the current directory, it does the following:
>
> * Finds the first parent directory containing `tf` files and adds `-chdir=$dir` to the command line arguments, to make Terraform change to that directory when it runs.
> * Updates the `TF_DATA_DIR` environment variable to make Terraform use the `.terraform` directory inside the current directory, next to the `tfvars` files rather than the configuration files.
> * Finds `tfvars` files in the current directory and updates the `TF_CLI_ARGS_plan` and `TF_CLI_ARGS_apply` environment variables to contain `-var-file=$filename` for each file. LTF follows Terraform's [ rules](https://www.terraform.io/language/values/variables#variable-definition-precedence) for which `tfvars` files to use.
> * Runs Terraform, passing along all command line arguments.
</details>

## Feature: Dynamic backends

LTF solves the [issue](https://github.com/hashicorp/terraform/issues/13022) of Terraform not supporting input variables in the backend configuration block. Note that LTF only adds support for input variables. It does not support accessing local values, nor data sources.

<details>
  <summary>How does it work?</summary>

> When LTF runs, it does the following:
>
> * Reads the backend block from the Terraform configuration.
> * Renders the backend block using HCL using Terraform variables.
> * Passes each line from the rendered backend block to Terraform using the `-backend-config=` command line argument, which takes precedence over the values in the file.
</details>

LTF also automatically uses backend configuration files from the current directory. This allows storing backend configuration in files alongside variables files.

<details>
  <summary>How does it work?</summary>

> When LTF runs and finds `tfbackend` files in the current directory, it does the following:
>
> * Finds `tfbackend` files in the current directory and updates the `TF_CLI_ARGS_init` environment variable to contain `-backend-config=$filename` for each file.
</details>

## Other ideas for later or never

* Creating backend resources.
  * Different backends have different methods for creation so this tool could only support some, if it supports any at all.
  * For S3/DynamoDB it might be better to create a separate tool or easy to use CloudFormation template that can be used to set up the Terraform project.
* Managing secrets
  * For SOPS, I'm not sure.
    * Consider adding a hooks system to run commands like SOPS and use the results as Terraform variables or environment variables.
    * Or have first-class support for running SOPS.
    * Or include the SOPS library and use it directly.
    * Or use the [Terraform Provider](https://github.com/carlpett/terraform-provider-sops)
* Creating backend resources and managing secrets in AWS.
  * Create S3/DynamoDB/KMS resources with CloudFormation.
  * Create built-in CLI and TUI interfaces for managing KMS secrets that get committed to git.
    * Examples: [SOPS](https://github.com/mozilla/sops) and [Tomb](https://github.com/gabrielfalcao/tomb)
  * Also consider the [aws_kms_secrets](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_secrets) data source.
* Remote modules, e.g. Terragrunt.
  * Probably will not implement because the project structure is too different.
* Hooks system.
  * Would be good to allow integrating SOPS, for example.
  * Might need to support a file in any location (current dir or parent dirs).
    * With tfvars
    * With config
    * In top of git repo
  * Maybe call it `ltf.yaml`
