variable "account_name" {
  description = "Name of AWS account"
  type        = string
}

variable "application_subnet_ids" {
  description = "Application subnet IDs in VPC"
  type        = list(string)
}

variable "environment_name" {
  description = "The name of the environment the fixtures container is deployed to"
  type        = string
}

variable "cloudwatch_kms_key_id" {
  description = "KMS key used to encrypt CloudWatch logs"
  type        = string
}

variable "ecr_image_uri" {
  description = "The URI of the image the container should use"
  type        = string
}

variable "public_subnet_ids" {
  description = "Public subnet IDs in VPC"
  type        = list(string)
}

variable "service_url" {
  description = "URL of the LPA Store service in this environment"
  type        = string
}

variable "vpc_id" {
  description = "ID of VPC the ECS Container will sit in"
  type        = string
}
