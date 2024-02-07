terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.15.0"
      configuration_aliases = [
        aws.backup-account,
        aws.source-account
      ]
    }
  }
  required_version = ">= 1.0.0"
}
