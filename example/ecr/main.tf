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
  backend "s3" {}
}

resource "random_id" "ecr" {
  byte_length = 4
  prefix      = var.env
}
