resource "aws_iam_role_policy_attachment" "cross_account_policy_attachment" {
  role       = var.s3_replication_role.name
  policy_arn = aws_iam_policy.cross_account_backup_policy.arn
  provider   = aws.source-account
}

resource "aws_iam_policy" "cross_account_backup_policy" {
  name        = "cross-account-s3-backup-policy-${var.environment_name}"
  description = "IAM Policy for s3 replication in ${var.environment_name}"
  policy      = data.aws_iam_policy_document.cross_account_policy.json
  provider    = aws.source-account
}

data "aws_iam_policy_document" "cross_account_policy" {
  provider = aws.source-account
  statement {
    sid    = "AllowReplication"
    effect = "Allow"
    actions = [
      "s3:ReplicateObject",
      "s3:ReplicateDelete",
      "s3:ReplicateTags",
      "s3:GetObjectVersionTagging",
      "s3:ObjectOwnerOverrideToBucketOwner"
    ]

    condition {
      test     = "StringLikeIfExists"
      variable = "s3:x-amz-server-side-encryption"
      values = [
        "aws:kms",
        "AES256"
      ]
    }

    condition {
      test     = "StringLikeIfExists"
      variable = "s3:x-amz-server-side-encryption-aws-kms-key-id"
      values = [
        aws_kms_key.key.arn
      ]
    }
    resources = ["${aws_s3_bucket.bucket.arn}/*"]
  }

  statement {
    sid    = "AllowEncryptCrossAccount"
    effect = "Allow"
    actions = [
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

      values = [
        "${aws_s3_bucket.bucket.arn}/*",
      ]
    }

    resources = [aws_kms_key.key.arn]
  }
}
