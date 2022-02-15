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
  default = "eu-west-1"
}

terraform {
  backend "s3" {}
}

resource "random_id" "iam" {
  byte_length = 4
  prefix      = var.env
}
