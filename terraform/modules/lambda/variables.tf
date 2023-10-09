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

variable "cloudwatch_kms_key_id" {
  description = "KMS key used to encrypt CloudWatch logs"
  type        = string
}

variable "environment_variables" {
  description = "A map that defines environment variables for the Lambda Function"
  type        = map(string)
  default     = {}
}
