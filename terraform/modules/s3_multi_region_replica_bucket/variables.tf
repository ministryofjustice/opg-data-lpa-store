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

variable "force_destroy" {
  type    = bool
  default = false
}

variable "kms_allowed_iam_roles" {
  default = []
  type    = list(string)
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

variable "s3_access_logging_bucket" {
  type = string
}

variable "s3_replication_role" {
  type = object({
    arn  = string
    name = string
  })
}
