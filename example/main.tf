terraform {
  backend "local" {
    // Put something invalid here to prevent init from working
    // except when it has been overridden by LTF.
    // This is only necessary because the local backend works with an empty
    // configuration, so it doesn't make the best example, but I don't want
    // to introduce credentials for other backend types just for an example.
    path = var.set_by_ltf
  }
}

variable "env" {
  type = string
}

variable "byte_length" {
  type = number
}

resource "random_id" "this" {
  byte_length = var.byte_length
  prefix      = "${var.env}-"
}
