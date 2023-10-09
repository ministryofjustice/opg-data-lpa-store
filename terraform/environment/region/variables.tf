variable "environment_name" {
  description = "The name of the environment the region is deployed to"
  type        = string
}

variable "app_version" {
  description = "Version of application to deploy"
  type        = string
}

variable "dynamodb_arn" {
  description = "ARN of DynamoDB global endpoint"
  type        = string
}
