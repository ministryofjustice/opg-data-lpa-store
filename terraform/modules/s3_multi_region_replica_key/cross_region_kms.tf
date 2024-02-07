resource "aws_kms_key" "eu_west_1" {
  description         = "Primary KMS Key for ${var.environment_name} S3 Replication"
  enable_key_rotation = true
  multi_region        = true
  policy              = data.aws_iam_policy_document.kms_policy.json
  provider            = aws.eu-west-1
}

resource "aws_kms_alias" "eu_west_1" {
  name          = "alias/${var.environment_name}/s3-replication-key"
  target_key_id = aws_kms_key.eu_west_1.id
  provider      = aws.eu-west-1
}

data "aws_iam_policy_document" "kms_policy" {
  statement {
    sid       = "Enable IAM User Permissions"
    effect    = "Allow"
    actions   = ["kms:*"]
    resources = ["*"]

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.eu_west_1.account_id}:root"]
    }
  }
  provider = aws.eu-west-1
}

resource "aws_kms_replica_key" "eu_west_2" {
  description             = "Replica KMS Key for ${var.environment_name} S3 Replication"
  deletion_window_in_days = 7
  primary_key_arn         = aws_kms_key.eu_west_1.arn
  policy                  = data.aws_iam_policy_document.kms_policy.json
  provider                = aws.eu-west-2
}

resource "aws_kms_alias" "eu_west_2" {
  name          = "alias/${var.environment_name}/s3-replication-key"
  target_key_id = aws_kms_replica_key.eu_west_2.id
  provider      = aws.eu-west-2
}
