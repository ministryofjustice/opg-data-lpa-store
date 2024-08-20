variable "accounts" {
  type = map(
    object({
      account_id                         = string
      account_name                       = string
      is_production                      = bool
      jwt_key_cross_account_access       = list(string)
      jwt_key_cross_account_access_roles = list(string)
    })
  )
}

locals {
  account_name = lower(replace(terraform.workspace, "_", "-"))
  account      = var.accounts[local.account_name]

  mandatory_moj_tags = {
    business-unit    = "OPG"
    application      = "opg-data-lpa-store"
    account          = local.account.account_name
    environment-name = local.account.account_name
    is-production    = local.account.is_production
    owner            = "opgteam@digital.justice.gov.uk"
  }

  optional_tags = {
    source-code            = "https://github.com/ministryofjustice/opg-data-lpa-store"
    infrastructure-support = "opgteam@digital.justice.gov.uk"
  }

  default_tags = merge(local.mandatory_moj_tags, local.optional_tags)
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
