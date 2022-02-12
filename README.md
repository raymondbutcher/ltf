# ltf

LTF is a lightweight, transparent wrapper for Terraform that adds only 2 features:

* Allow variables in backend configuration blocks.
* Automatically use configuration from parent directories. (not implemented)

Because LTF is a transparent wrapper, all standard Terraform command line options can be used when using LTF.

## Status

This project is just an idea that has not been implemented.

## Installation

LTF is released as a single binary. Download and add it to your `PATH`. Make it executable if needed.

## Usage

Run `ltf` instead of `terraform`.

Example:

```
$ ltf init
$ ltf plan
$ ltf apply -target=random_id.this
```

## Feature: allow variables in backend configuration blocks

LTF solves the [issue](https://github.com/hashicorp/terraform/issues/13022) of Terraform not supporting input variables in the backend configuration block. Note that LTF only adds support for input variables. It does not support accessing local values, nor data sources.

<details>
  <summary>How does it work?</summary>

> When LTF runs, it does the following:
>
> * Reads the backend block from the Terraform configuration.
> * Renders the backend block using HCL using Terraform variables.
> * Passes each line from the rendered backend block to Terraform using the `-backend-config=` command line argument, which takes precedence over the values in the file.
</details>

## Feature: automatically use configuration from parent directories

LTF makes it easy to work with multiple environments. It does this by using variables from the current directory with configuration from a parent directory.

This allows for the following project structure:

```
terraform
├── ecr
│   ├── dev
│   │   └── ecr.dev.auto.tfvars
│   ├── live
│   │   ├── eu-central-1
│   │   │   └── ecr.live.eu-central-1.auto.tfvars
│   │   └── eu-west-1
│   │       └── ecr.live.eu-west-1.auto.tfvars
│   └── main.tf
└── iam
    ├── dev
    │   └── iam.dev.auto.tfvars
    ├── live
    │   └── iam.live.auto.tfvars
    └── main.tf
```

Usage looks like this:

```
$ cd ecr/dev
$ ltf plan
$ cd ../live/eu-west-1
$ ltf plan
```

<details>
  <summary>How does it work?</summary>

> When LTF runs and finds `tfvars` files in the current directory, but no `tf` files, then it does the following:
>
> * Uses the current directory as the *variables directory*.
> * Finds the first parent directory containing `tf` files and uses it as the *configuration directory*.
> * Reads `tfvars` files from the *variables directory* and exports matching `TF_VAR_name` environment variables.
> * Exports the `TF_DATA_DIR=$variablesdir/.terraform` environment variable so that Terrafor places the `.terraform` directory in the *variables directory*.
> * Runs `terraform -chdir=$configurationdir ...`
</details>

## Other ideas for later

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
