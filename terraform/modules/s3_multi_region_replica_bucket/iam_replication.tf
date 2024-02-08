resource "aws_iam_role_policy_attachment" "replication_role_s3_permissions" {
  role       = var.s3_replication_role.name
  policy_arn = aws_iam_policy.replication_role_s3_permissions.arn
}

resource "aws_iam_policy" "replication_role_s3_permissions" {
  name        = "s3-replication-policy-${var.bucket_name}"
  description = "S3 Replication Policy for ${var.bucket_name}"
  policy      = data.aws_iam_policy_document.replication_role_s3_permissions.json
}

data "aws_iam_policy_document" "replication_role_s3_permissions" {

  statement {
    sid    = "AllowReplicationConfiguration"
    effect = "Allow"
    actions = [
      "s3:ListBucket",
      "s3:GetReplicationConfiguration",
      "s3:GetObjectVersionForReplication",
      "s3:GetObjectVersionAcl",
      "s3:GetObjectVersionTagging",
      "s3:GetObjectRetention",
      "s3:GetObjectLegalHold"
    ]
    resources = [
      "${aws_s3_bucket.bucket.arn}/*",
      aws_s3_bucket.bucket.arn
    ]
  }
  statement {
    sid    = "AllowCrossRegionReplication"
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
      values   = concat(var.replication_kms_key_arns, [aws_kms_key.s3.arn])
    }
    resources = ["${aws_s3_bucket.bucket.arn}/*"]
  }
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
    resources = [
      aws_kms_key.s3.arn
    ]
  }
}
