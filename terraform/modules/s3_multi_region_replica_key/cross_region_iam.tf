resource "aws_iam_role_policy_attachment" "replication_role_kms_permissions" {
  role       = var.s3_replication_role.name
  policy_arn = aws_iam_policy.replication_role_kms_permissions.arn
}

resource "aws_iam_policy" "replication_role_kms_permissions" {
  name        = "s3-regional-replication-kms-policy-${var.environment_name}"
  description = "S3 Regional Replication KMS Policy for ${var.environment_name}"
  policy      = data.aws_iam_policy_document.replication_role_kms_permissions.json
}

data "aws_iam_policy_document" "replication_role_kms_permissions" {
  statement {
    sid    = "AllowKeysEncryptDecryptForS3CrossRegion"
    effect = "Allow"
    actions = [
      "kms:Decrypt",
      "kms:Encrypt"
    ]

    condition {
      test     = "StringLike"
      variable = "kms:ViaService"

      values = [
        "s3.eu-west-1.amazonaws.com",
        "s3.eu-west-2.amazonaws.com",
      ]
    }

    condition {
      test     = "StringLike"
      variable = "kms:EncryptionContext:aws:s3:arn"

      values = [for bucket_arn in var.bucket_arns : "${bucket_arn}/*"]
    }

    resources = [
      aws_kms_key.eu_west_1.arn,
      aws_kms_replica_key.eu_west_2.arn
    ]
  }
}
