provider "aws" {}

data "ns_connection" "test" {
  type = "fargate/service"
}

variable "var1" {
  type        = string
  description = "Description for var1"
  default     = "some-value"
}

variable "var2" {
  type        = list(string)
  description = "Description for var2"
  default     = ["list", "of", "values"]
}

output "some_value" {
  value       = module.network.vpc_id
  description = "string ||| Some value"
}

output "sensitive_value" {
  value       = module.network.vpc_id
  description = "string ||| "
  sensitive   = true
}

output "list_value" {
  value       = module.network.vpc_id
  description = "list(string) ||| "
}
