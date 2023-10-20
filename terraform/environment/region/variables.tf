variable "environment_name" {
  description = "The name of the environment the region is deployed to"
  type        = string
}

variable "app_version" {
  description = "Version of application to deploy"
  type        = string
}

variable "dynamodb_arn" {
  description = "ARN of DynamoDB table"
  type        = string
}

variable "dynamodb_name" {
  description = "Name of DynamoDB table"
  type        = string
}

variable "allowed_arns" {
  description = "List of external ARNs allowed to access the API Gateway"
  type        = list(string)
}

locals {
  policy_region_prefix = lower(replace(data.aws_region.current.name, "-", ""))
}
