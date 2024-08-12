module "jwt_kms" {
  source                  = "../modules/kms_key"
  encrypted_resource      = "jwt key secret"
  kms_key_alias_name      = "${data.aws_default_tags.default.tags.application}/${data.aws_default_tags.default.tags.account}/jwt-key"
  enable_key_rotation     = true
  enable_multi_region     = true
  deletion_window_in_days = 10
  kms_key_policy          = data.aws_default_tags.default.tags.account == "development" ? data.aws_iam_policy_document.jwt_kms_merged.json : data.aws_iam_policy_document.jwt_kms.json
  providers = {
    aws.eu_west_1 = aws.management_eu_west_1
    aws.eu_west_2 = aws.management_eu_west_2
  }
}
data "aws_iam_policy_document" "jwt_kms_merged" {
  provider = aws.global
  source_policy_documents = [
    data.aws_iam_policy_document.jwt_kms.json,
    data.aws_iam_policy_document.jwt_kms_development_account_operator_admin.json
  ]
}

data "aws_iam_policy_document" "jwt_kms" {
  provider = aws.global

  statement {
    sid    = "Enable IAM User Permissions"
    effect = "Allow"
    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.management.account_id}:root"]
    }
    actions = [
      "kms:*",
    ]
    resources = [
      "*",
    ]
  }

  statement {
    sid    = "Allow Key to be used for Encryption"
    effect = "Allow"
    resources = [
      "*"
    ]
    actions = [
      "kms:Encrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:DescribeKey",
    ]

    principals {
      type = "AWS"
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.management.account_id}:role/breakglass",
      ]
    }
  }

  statement {
    sid    = "Cross account access"
    effect = "Allow"
    resources = [
      "*"
    ]
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey*",
      "kms:DescribeKey",
    ]

    principals {
      type        = "AWS"
      identifiers = local.account.jwt_key_cross_account_access
    }
    condition {
      test     = "ArnLike"
      variable = "aws:PrincipalArn"
      values   = local.account.jwt_key_cross_account_access_roles
    }

  }

  statement {
    sid    = "Key Administrator"
    effect = "Allow"
    resources = [
      "*"
    ]
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
      "kms:CancelKeyDeletion",
      "kms:ReplicateKey"
    ]

    principals {
      type = "AWS"
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.management.account_id}:role/breakglass",
        "arn:aws:iam::${data.aws_caller_identity.management.account_id}:role/lpa-store-ci",
        "arn:aws:iam::${data.aws_caller_identity.management.account_id}:role/modernising-lpa-ci",
      ]
    }
  }
}

data "aws_iam_policy_document" "jwt_kms_development_account_operator_admin" {
  provider = aws.global
  statement {
    sid    = "Dev Account Key Administrator"
    effect = "Allow"
    resources = [
      "*"
    ]
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
      type = "AWS"
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.management.account_id}:role/operator"
      ]
    }
  }
}

