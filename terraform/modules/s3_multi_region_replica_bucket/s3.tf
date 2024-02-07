resource "aws_s3_bucket" "bucket" {
  bucket        = var.bucket_name
  force_destroy = var.force_destroy
}

resource "aws_s3_bucket_ownership_controls" "bucket_object_ownership" {
  bucket = aws_s3_bucket.bucket.id
  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_versioning" "bucket_versioning" {
  bucket = aws_s3_bucket.bucket.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "bucket_encryption_configuration" {
  bucket = aws_s3_bucket.bucket.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.s3.id
      sse_algorithm     = "aws:kms"
    }
  }
}

# Clarify if we want this
resource "aws_s3_bucket_lifecycle_configuration" "bucket" {
  depends_on = [aws_s3_bucket_versioning.bucket_versioning]
  bucket     = aws_s3_bucket.bucket.id

  rule {
    id     = "ExpireNonCurrentVersionsAfter${var.expire_non_current_object_version_days}Days"
    status = "Enabled"

    expiration {
      expired_object_delete_marker = true
    }

    noncurrent_version_expiration {
      noncurrent_days = var.expire_non_current_object_version_days
    }
  }
}

resource "aws_s3_bucket_replication_configuration" "bucket_replication" {
  depends_on = [aws_s3_bucket_versioning.bucket_versioning]

  role   = var.s3_replication_role.arn
  bucket = aws_s3_bucket.bucket.id

  dynamic "rule" {
    for_each = var.replication_configuration
    content {
      id       = "ReplicationTo${rule.value["bucket"].id}"
      priority = rule.key
      status   = "Enabled"


      destination {
        account = rule.value["account_id"]
        bucket  = rule.value["bucket"].arn

        encryption_configuration {
          replica_kms_key_id = rule.value["kms_key_arn"]
        }

        access_control_translation {
          owner = "Destination"
        }
      }

      delete_marker_replication {
        status = "Enabled"
      }

      filter {}

      source_selection_criteria {
        sse_kms_encrypted_objects {
          status = "Enabled"
        }
      }
    }
  }
}

resource "aws_s3_bucket_public_access_block" "public_access_policy" {
  bucket = aws_s3_bucket.bucket.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# resource "aws_s3_bucket_logging" "bucket" {
#   bucket = aws_s3_bucket.bucket.id

#   target_bucket = var.s3_access_logging_bucket
#   target_prefix = "log/${aws_s3_bucket.bucket.id}/"
# }

resource "aws_s3_bucket_policy" "bucket" {
  depends_on = [aws_s3_bucket_public_access_block.public_access_policy]
  bucket     = aws_s3_bucket.bucket.id
  policy     = local.cross_account_read ? data.aws_iam_policy_document.cross_account_read[0].json : data.aws_iam_policy_document.bucket_default.json
}

data "aws_iam_policy_document" "bucket_default" {
  policy_id = "PutObjPolicy"

  statement {
    sid       = "DenyUnEncryptedObjectUploads"
    effect    = "Deny"
    actions   = ["s3:PutObject"]
    resources = ["${aws_s3_bucket.bucket.arn}/*"]

    condition {
      test     = "StringNotEquals"
      variable = "s3:x-amz-server-side-encryption"
      values   = ["aws:kms"]
    }

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }

  statement {
    sid     = "DenyNoneSSLRequests"
    effect  = "Deny"
    actions = ["s3:*"]
    resources = [
      aws_s3_bucket.bucket.arn,
      "${aws_s3_bucket.bucket.arn}/*"
    ]

    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = [false]
    }

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }
}

data "aws_iam_policy_document" "cross_account_read" {
  count                   = local.cross_account_read ? 1 : 0
  source_policy_documents = [data.aws_iam_policy_document.bucket_default.json]
  statement {
    sid    = "DelegateS3Access"
    effect = "Allow"
    actions = [
      "s3:ListBucket",
      "s3:GetObject",
      "s3:GetObjectTagging",
      "s3:GetObjectVersionTagging"
    ]

    principals {
      type        = "AWS"
      identifiers = [for account in var.accounts_allowed_to_read : "arn:aws:iam::${account}:root"]
    }

    resources = [
      aws_s3_bucket.bucket.arn,
      "${aws_s3_bucket.bucket.arn}/*"
    ]
  }
}
