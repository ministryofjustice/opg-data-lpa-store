terraform {
  required_version = ">= 1.4.0"

  required_providers {
    aws = {
      source = "hashicorp/aws"
      configuration_aliases = [
        aws.region,
        aws.management,
      ]
    }
  }
}

data "aws_region" "current" {
  provider = aws.region
}
