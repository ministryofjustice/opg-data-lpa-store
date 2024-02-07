locals {
  cross_account_read = length(var.accounts_allowed_to_read) != 0 ? true : false
}

variable "accounts_allowed_to_read" {
  default = []
  type    = list(string)
}

variable "bucket_name" {
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

variable "replication_configuration" {
  default = []
  type = list(object({
    account_id = string
    bucket = object({
      arn = string
      id  = string
    })
    kms_key_arn = string
  }))
}

variable "replication_kms_key_arns" {
  type = list(string)
}

# variable "s3_access_logging_bucket" {
#   type = string
# }

variable "s3_replication_role" {
  type = object({
    arn  = string
    name = string
  })
}
