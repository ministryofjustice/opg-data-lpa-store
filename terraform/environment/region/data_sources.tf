data "aws_region" "current" {
  provider = aws.region
}

data "aws_caller_identity" "current" {
  provider = aws.region
}

data "aws_default_tags" "default" {
  provider = aws.region
}

data "aws_vpc" "main" {
  filter {
    name   = "tag:name"
    values = ["opg-data-lpa-store-${var.environment.account_name}-vpc"]
  }

  filter {
    name   = "tag:Name"
    values = ["opg-data-lpa-store-${var.environment.account_name}-vpc"]
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
  name     = "${data.aws_default_tags.default.tags.application}/${data.aws_default_tags.default.tags.account}/jwt-key"
  provider = aws.management
}

data "aws_kms_alias" "jwt_key" {
  name     = "alias/${data.aws_default_tags.default.tags.application}/${data.aws_default_tags.default.tags.account}/jwt-key"
  provider = aws.management
}
