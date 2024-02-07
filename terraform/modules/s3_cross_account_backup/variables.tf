locals {
  s3_logging_bucket_name = "${var.s3_access_logging_bucket_prefix}-${data.aws_region.backup_account.name}"
}

variable "bucket_name" {
  type = string
}

variable "environment_name" {
  type = string
}

variable "expire_non_current_object_version_days" {
  default     = 90
  description = "How Many days to keep non current versions of objects for."
  type        = number
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
