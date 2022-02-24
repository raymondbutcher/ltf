terraform {
  backend "local" {}
}

variable "byte_length" {
  type = number
}

variable "color" {
  type = string
  default = ""
}

variable "env" {
  type = string
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
