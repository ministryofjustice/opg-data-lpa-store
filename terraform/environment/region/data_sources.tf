data "aws_region" "current" {
  provider = aws.region
}

data "aws_caller_identity" "current" {
  provider = aws.region
}

# we could use this data source instead of using an input variable for the account name
# data "aws_default_tags" "current" {
#   provider = aws.region
# }


data "aws_vpc" "main" {
  filter {
    name   = "tag:name"
    values = ["opg-data-lpa-store-${var.account_name}-vpc"]
  }

  provider = aws.region
}

data "aws_subnets" "public" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.main.id]
  }

  filter {
    name   = "tag:Name"
    values = ["public-*"]
  }

  provider = aws.region
}

data "aws_subnets" "application" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.main.id]
  }

  filter {
    name   = "tag:Name"
    values = ["application-*"]
  }

  provider = aws.region
}

data "aws_secretsmanager_secret" "jwt_secret_key" {
  name     = "${var.account_name}/jwt-key"
  provider = aws.region
}
