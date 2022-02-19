terraform {
  backend "local" {}
}

variable "env" {
  type = string
}

variable "byte_length" {
  type = number
}

variable "secrets" {
  type = string
}

resource "random_id" "this" {
  byte_length = var.byte_length
  prefix      = "${var.env}-"
}

output "secrets" {
  value = var.secrets
}
