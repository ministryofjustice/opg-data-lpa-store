locals {
  s3_logging_bucket_name = "${var.s3_access_logging_bucket_prefix}-${data.aws_region.backup_account.name}"
}

variable "bucket_name" {
  type = string
}

variable "environment_name" {
  type = string
}

variable "force_destroy" {
  type    = bool
  default = false
}

variable "s3_access_logging_bucket_prefix" {
  type = string
}

variable "s3_replication_role" {
  type = object({
    name = string
  })
}
