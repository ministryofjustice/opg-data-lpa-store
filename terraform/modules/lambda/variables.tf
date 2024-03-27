variable "account_name" {
  description = "Name of AWS account"
  type        = string
}

variable "environment_name" {
  description = "The name of the environment the lambda is deployed to"
  type        = string
}

variable "lambda_name" {
  description = "The name of the lambda function"
  type        = string
}

variable "ecr_image_uri" {
  description = "The URI of the image lambda should use"
  type        = string
}

variable "event_bus_arn" {
  description = "The ARN of the event bus to send update events to"
  type        = string
}

variable "cloudwatch_kms_key_id" {
  description = "KMS key used to encrypt CloudWatch logs"
  type        = string
}

variable "environment_variables" {
  description = "A map that defines environment variables for the Lambda Function"
  type        = map(string)
  default     = {}
}

variable "subnet_ids" {
  description = "IDs of the subnets the Lambda Function will sit in"
  type        = list(string)
}

variable "vpc_id" {
  description = "ID of VPC the Lambda Function will sit in"
  type        = string
}
