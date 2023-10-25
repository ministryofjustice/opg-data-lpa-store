terraform {
  backend "s3" {
    bucket         = "opg.terraform.state"
    key            = "opg-data-lpa-deed-account/terraform.tfstate"
    encrypt        = true
    region         = "eu-west-1"
    role_arn       = "arn:aws:iam::311462405659:role/lpa-store-ci"
    dynamodb_table = "remote_lock"
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.8.0"
    }
  }
  required_version = ">= 1.4.0"
}

provider "aws" {
  alias  = "eu_west_1"
  region = "eu-west-1"

  assume_role {
    role_arn     = "arn:aws:iam::${local.account.account_id}:role/${var.default_role}"
    session_name = "terraform-session"
  }

  default_tags {
    tags = local.default_tags
  }
}

provider "aws" {
  alias  = "eu_west_2"
  region = "eu-west-2"

  assume_role {
    role_arn     = "arn:aws:iam::${local.account.account_id}:role/${var.default_role}"
    session_name = "terraform-session"
  }

  default_tags {
    tags = local.default_tags
  }
}

provider "aws" {
  alias  = "global"
  region = "us-east-1"

  assume_role {
    role_arn     = "arn:aws:iam::${local.account.account_id}:role/${var.default_role}"
    session_name = "terraform-session"
  }

  default_tags {
    tags = local.default_tags
  }
}

provider "aws" {
  alias  = "management_eu_west_1"
  region = "eu-west-1"

  assume_role {
    role_arn     = "arn:aws:iam::311462405659:role/${var.management_role}"
    session_name = "terraform-session"
  }

  default_tags {
    tags = local.default_tags
  }
}

provider "aws" {
  alias  = "management_eu_west_2"
  region = "eu-west-2"

  assume_role {
    role_arn     = "arn:aws:iam::311462405659:role/${var.management_role}"
    session_name = "terraform-session"
  }

  default_tags {
    tags = local.default_tags
  }
}
