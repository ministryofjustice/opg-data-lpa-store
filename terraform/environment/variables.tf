locals {
  backup_account_id = 238302996107
  environment_name  = lower(replace(terraform.workspace, "_", "-"))
  environment       = contains(keys(var.environments), local.environment_name) ? var.environments[local.environment_name] : var.environments["default"]

  is_ephemeral = !contains(keys(var.environments), local.environment_name)

  cross_account_backup_enabled = !local.is_ephemeral

  default_tags = merge(local.mandatory_moj_tags, local.optional_tags)
  mandatory_moj_tags = {
    business-unit    = "OPG"
    application      = "opg-data-lpa-store"
    environment-name = local.environment_name
    account          = local.environment.account_name
    is-production    = local.environment.is_production
    owner            = "opgteam@digital.justice.gov.uk"
  }

  optional_tags = {
    source-code            = "https://github.com/ministryofjustice/opg-data-lpa-store"
    infrastructure-support = "opgteam@digital.justice.gov.uk"
  }
}

variable "environments" {
  type = map(
    object({
      account_id    = string
      account_name  = string
      is_production = bool
      allowed_arns  = list(string)
    })
  )
}

variable "default_role" {
  description = "Role to assume in LPA Store account"
  type        = string
  default     = "lpa-store-ci"
}

variable "management_role" {
  description = "Role to assume in Management account"
  type        = string
  default     = "lpa-store-ci"
}

variable "app_version" {
  description = "Version of application to deploy"
  type        = string
  default     = "latest"
}
