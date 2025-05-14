variable "app_version" {
  description = "Version of application to deploy"
  type        = string
}

variable "environment" {
  type = object({
    account_id            = string
    account_name          = string
    allowed_arns          = list(string)
    allowed_wildcard_arns = optional(list(string), [])
  })
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
  type        = any
}

variable "has_fixtures" {
  description = "Whether the environment should have a fixtures container"
  type        = bool
  default     = false
}

variable "lpa_store_static_bucket" {
  description = "LPA Store Static bucket object for the region"
  type        = any
}

variable "lpa_store_static_bucket_kms_key" {
  description = "LPA Store Static bucket KMS Key object for the region"
  type        = any
}
