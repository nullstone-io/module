provider "aws" {
  alias = "test"
}

variable "var1" {
  type        = string
  description = "Description for var1"
  default     = "some-value"
}

module "sub" {
  source = "./sub"

  var2 = var.var1
}
