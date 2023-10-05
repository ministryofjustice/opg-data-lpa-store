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
