data "aws_caller_identity" "backup_account" {
  provider = aws.backup-account
}

data "aws_caller_identity" "source_account" {
  provider = aws.source-account
}

data "aws_region" "backup_account" {
  provider = aws.backup-account
}
