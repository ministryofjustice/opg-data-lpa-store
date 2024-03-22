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

variable "service_url" {
  description = "URL of the LPA Store service in this environment"
  type        = string
}

variable "subnet_ids" {
  description = "IDs of the subnets the ECS Container will sit in"
  type        = list(string)
}

variable "vpc_id" {
  description = "ID of VPC the ECS Container will sit in"
  type        = string
}
