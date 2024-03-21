data "aws_ecr_repository" "fixtures" {
  name     = "lpa-store/fixtures"
  provider = aws.management
}
