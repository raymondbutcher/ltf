variable "app" {
  default = "iam"
}

variable "env" {
  type = string
}

variable "aws_profile" {
  type = string
}

variable "aws_region" {
  type = string
}

terraform {
  backend "s3" {
    bucket  = "ltf-project"
    key     = "example/${var.app}/${var.env}/${var.aws_region}/terraform.tfstate"
    region  = var.region
    profile = var.aws_profile
  }
}

resource "random_id" "ecr" {
  byte_length = 4
  prefix      = var.env
}
