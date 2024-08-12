variable "is_production" {
  description = "Whether this is a production environment"
  type        = bool
}

variable "vpc_cidr" {
  default     = "10.0.0.0/16"
  description = "CIDR Range for the VPC"
  type        = string
}
