terraform {
  required_version = ">= 1.4.0"

  required_providers {
    aws = {
      version = ">= 5.8.0"
      source  = "hashicorp/aws"
      configuration_aliases = [
        aws.global,
        aws.management,
        aws.region,
      ]
    }
  }
}
