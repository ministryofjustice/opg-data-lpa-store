variable "bucket_arns" {
  type = list(string)
}

variable "environment_name" {
  type = string
}

variable "s3_replication_role" {
  type = object({
    name = string
  })
}
