variable "environment_name" {
  description = "The name of the environment the fixtures container is deployed to"
  type        = string
}

variable "cloudwatch_kms_key_id" {
  description = "KMS key used to encrypt CloudWatch logs"
  type        = string
}

variable "service_url" {
  description = "URL of the LPA Store service in this environment"
  type        = string
}
