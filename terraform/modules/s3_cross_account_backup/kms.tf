resource "aws_kms_alias" "alias" {
  name          = "alias/${var.environment_name}/s3-replication-key"
  target_key_id = aws_kms_key.key.id
  provider      = aws.backup-account
}

resource "aws_kms_key" "key" {
  description         = "KMS Key for ${var.environment_name} cross account S3 replication"
  enable_key_rotation = true
  policy              = data.aws_iam_policy_document.key.json
  provider            = aws.backup-account
}

data "aws_iam_policy_document" "key" {
  provider = aws.backup-account
  statement {
    sid       = "Enable IAM User Permissions"
    effect    = "Allow"
    actions   = ["kms:*"]
    resources = ["*"]

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.backup_account.account_id}:root"]
    }
  }

  statement {
    sid       = "Enable cross account encrypt access for S3 Cross Region Replication"
    effect    = "Allow"
    actions   = ["kms:Encrypt"]
    resources = ["*"]

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.source_account.account_id}:root"]
    }
  }
}
