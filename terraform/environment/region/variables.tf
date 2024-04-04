variable "allowed_arns" {
  description = "List of external ARNs allowed to access the API Gateway"
  type        = list(string)
}

variable "allowed_wildcard_arns" {
  description = "List of wildcard-containing external ARNs allowed to access the API Gateway"
  type        = list(string)
}

variable "account_name" {
  description = "Name of AWS account"
  type        = string
}

variable "app_version" {
  description = "Version of application to deploy"
  type        = string
}

variable "dns_weighting" {
  description = "What percentage of DNS traffic to send to this region"
  type        = number
  default     = 50
}

variable "dynamodb_arn" {
  description = "ARN of DynamoDB table"
  type        = string
}

variable "dynamodb_name" {
  description = "Name of DynamoDB table"
  type        = string
}

variable "dynamodb_arn_changes" {
  description = "ARN of DynamoDB table for changes"
  type        = string
}

variable "dynamodb_name_changes" {
  description = "Name of DynamoDB table for changes"
  type        = string
}

variable "environment_name" {
  description = "The name of the environment the region is deployed to"
  type        = string
}

variable "event_bus" {
  description = "Event bus to send events to"
}

variable "has_fixtures" {
  description = "Whether the environment should have a fixtures container"
  type        = bool
  default     = false
}

variable "lpa_store_static_bucket" {
  description = "LPA Store Static bucket object for the region"
}

variable "lpa_store_static_bucket_kms_key" {
  description = "LPA Store Static bucket KMS Key object for the region"
}
