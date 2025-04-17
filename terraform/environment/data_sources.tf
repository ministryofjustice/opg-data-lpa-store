data "aws_caller_identity" "current" {
  provider = aws.eu_west_1
}

data "aws_region" "current" {
  provider = aws.eu_west_1
}

data "aws_default_tags" "current" {
  provider = aws.eu_west_1
}
