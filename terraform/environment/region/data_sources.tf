data "aws_vpc" "main" {
  filter {
    name   = "tag:name"
    values = ["opg-data-lpa-store-${var.account_name}-vpc"]
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
