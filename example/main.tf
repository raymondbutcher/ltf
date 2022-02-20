terraform {
  backend "local" {}
}

variable "env" {
  type = string
}

variable "byte_length" {
  type = number
}

variable "hook" {
  type = string
}

resource "random_id" "this" {
  byte_length = var.byte_length
  prefix      = "${var.env}-"
}

output "hook" {
  value = var.hook
}
