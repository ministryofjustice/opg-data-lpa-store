terraform {
  required_version = ">= 1.4.0"

  required_providers {
    aws = {
      source = "hashicorp/aws"
      configuration_aliases = [
        aws.global,
        aws.management,
        aws.region,
      ]
    }
  }
}

data "aws_region" "current" {
  provider = aws.region
}

data "aws_caller_identity" "current" {
  provider = aws.region
}

data "aws_default_tags" "current" {
  provider = aws.region
}
