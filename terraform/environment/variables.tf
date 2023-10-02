locals {
  environment_name = lower(replace(terraform.workspace, "_", "-"))
  environment      = contains(keys(var.environments), local.environment_name) ? var.environments[local.environment_name] : var.environments["default"]

  default_tags = merge(local.mandatory_moj_tags, local.optional_tags)
  mandatory_moj_tags = {
    business-unit    = "OPG"
    application      = "opg-data-lpa-deed"
    environment-name = local.environment_name
    account          = local.environment.account_name
    is-production    = local.environment.is_production
    owner            = "opgteam@digital.justice.gov.uk"
  }

  optional_tags = {
    source-code            = "https://github.com/ministryofjustice/opg-data-lpa-deed"
    infrastructure-support = "opgteam@digital.justice.gov.uk"
  }
}

variable "environments" {
  type = map(
    object({
      account_id    = string
      account_name  = string
      is_production = bool
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
