provider "aws" {}
data "aws_availability_zone" "this" {}

data "ns_connection" "service" {
  name     = "service"
  contract = "app/aws/ecs"
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
