variable "bool_default_value" {
  type    = bool
  default = true
}

variable "bool_no_value" { type = bool }

variable "bool_value" { type = bool }

###

variable "list_default_value" {
  type    = list(bool)
  default = [true]
}

variable "list_no_value" { type = list(bool) }

variable "list_value" { type = list(bool) }

###

variable "string_default_value" {
  type    = string
  default = "string_default_value"
}

variable "string_no_value" { type = string }

variable "string_value" { type = string }

###

variable "untyped_no_value" {}

variable "untyped_default_list_value" {
  default = ["untyped_default_list_value"]
}

variable "untyped_bool_value" {}

variable "untyped_string_value" {}
