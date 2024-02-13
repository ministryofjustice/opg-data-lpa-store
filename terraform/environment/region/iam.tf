moved {
  from = aws_iam_role_policy.lambda
  to   = aws_iam_role_policy.lambda_dynamodb
}

resource "aws_iam_role_policy" "lambda_dynamodb" {
  for_each = local.functions
  name     = "LambdaAllowDynamoDB"
  role     = module.lambda[each.key].iam_role.id
  policy   = data.aws_iam_policy_document.lambda_dynamodb_policy.json
  provider = aws.region
}

data "aws_iam_policy_document" "lambda_dynamodb_policy" {
  statement {
    sid       = "allowDynamoDB"
    effect    = "Allow"
    resources = [var.dynamodb_arn, var.dynamodb_arn_changes]
    actions = [
      "dynamodb:PutItem",
      "dynamodb:GetItem",
    ]
  }
}

resource "aws_iam_role_policy" "lambda_s3" {
  for_each = local.functions
  name     = "LambdaAllowS3"
  role     = module.lambda[each.key].iam_role.id
  policy   = data.aws_iam_policy_document.lambda_s3_policy.json
  provider = aws.region
}

data "aws_iam_policy_document" "lambda_s3_policy" {
  statement {
    sid    = "allowS3Access"
    effect = "Allow"
    resources = [
      var.lpa_store_static_bucket.arn,
      "${var.lpa_store_static_bucket.arn}/*",
    ]
    actions = [
      "s3:PutObject",
    ]
  }
  statement {
    sid       = "allowS3KMS"
    effect    = "Allow"
    resources = [var.lpa_store_static_bucket_kms_key.arn]
    actions = [
      "kms:GenerateDataKey",
      "kms:Encrypt"
    ]

    condition {
      test     = "StringLike"
      variable = "kms:ViaService"
      values   = ["s3.${data.aws_region.current.name}.amazonaws.com"]
    }

    condition {
      test     = "StringLike"
      variable = "kms:EncryptionContext:aws:s3:arn"
      values = [
        "${var.lpa_store_static_bucket.arn}/*",
      ]
    }
  }
}
