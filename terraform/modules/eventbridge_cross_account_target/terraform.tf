terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.15.0"
    }
  }
  required_version = ">= 1.0.0"
}

data "aws_region" "current" {}
