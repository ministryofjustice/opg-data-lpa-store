terraform {
  required_version = ">= 1.4.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
      configuration_aliases = [
        aws.global,
        aws.management,
        aws.region,
      ]
    }
  }
}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {
  provider = aws.region
}
