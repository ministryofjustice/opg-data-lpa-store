resource "aws_kms_key" "s3" {
  description             = "KMS Key for Encryption at Rest for S3 Bucket ${var.bucket_name}"
  deletion_window_in_days = 10
  enable_key_rotation     = true
  policy                  = data.aws_iam_policy_document.audit_kms.json
}

resource "aws_kms_alias" "audit_alias" {
  name          = "alias/S3-Encryption-${vr.bucket_name}"
  target_key_id = aws_kms_key.s3.key_id
}

# See the following link for further information
# https://docs.aws.amazon.com/kms/latest/developerguide/key-policies.html
data "aws_iam_policy_document" "s3_kms" {
  statement {
    sid       = "Enable Root account permissions on Key"
    effect    = "Allow"
    actions   = ["kms:*"]
    resources = ["*"]

    principals {
      type = "AWS"
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root",
      ]
    }
  }

  #   statement {
  #     sid       = "Allow Key to be used for Encryption"
  #     effect    = "Allow"
  #     resources = ["*"]
  #     actions = [
  #       "kms:Encrypt",
  #       "kms:Decrypt",
  #       "kms:ReEncrypt*",
  #       "kms:GenerateDataKey*",
  #       "kms:DescribeKey",
  #     ]

  #     principals {
  #       type = "Service"
  #       identifiers = [
  #         "rds.amazonaws.com",
  #         "firehose.amazonaws.com",
  #         "lambda.amazonaws.com"
  #       ]
  #     }
  #   }

  #   statement {
  #     sid       = "Allow Key to be used for Encryption by Lambda"
  #     effect    = "Allow"
  #     resources = ["*"]
  #     actions = [
  #       "kms:Encrypt",
  #       "kms:Decrypt",
  #       "kms:ReEncrypt*",
  #       "kms:GenerateDataKey*",
  #       "kms:DescribeKey",
  #     ]

  #     principals {
  #       type        = "AWS"
  #       identifiers = [var.audit_iam_role.arn]
  #     }
  #   }

  statement {
    sid       = "Key Administrator"
    effect    = "Allow"
    resources = ["*"]
    actions = [
      "kms:Create*",
      "kms:Describe*",
      "kms:Enable*",
      "kms:List*",
      "kms:Put*",
      "kms:Update*",
      "kms:Revoke*",
      "kms:Disable*",
      "kms:Get*",
      "kms:Delete*",
      "kms:TagResource",
      "kms:UntagResource",
      "kms:ScheduleKeyDeletion",
      "kms:CancelKeyDeletion"
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/breakglass"]
    }
  }
}
