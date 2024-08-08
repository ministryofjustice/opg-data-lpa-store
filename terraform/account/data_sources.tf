data "aws_default_tags" "default" {
  provider = aws.eu_west_1
}

data "aws_region" "eu_west_2" {
  provider = aws.eu_west_2
}

data "aws_caller_identity" "management" {
  provider = aws.management_eu_west_1
}

data "aws_caller_identity" "current" {
  provider = aws.eu_west_1
}
