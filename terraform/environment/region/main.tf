locals {
  functions = toset([
    "create",
    "get",
    "update",
  ])
}

module "lambda" {
  for_each = local.functions
  source   = "../../modules/lambda"

  environment_name      = var.environment_name
  lambda_name           = each.key
  ecr_image_uri         = "${data.aws_ecr_repository.lambda[each.key].repository_url}:${var.app_version}"
  cloudwatch_kms_key_id = aws_kms_key.cloudwatch.arn

  environment_variables = {
    DDB_TABLE_NAME_DEEDS   = var.dynamodb_name
    DDB_TABLE_NAME_CHANGES = var.dynamodb_name_changes
    JWT_SECRET_KEY         = "secret"
  }

  providers = {
    aws = aws.region
  }
}

data "aws_ecr_repository" "lambda" {
  for_each = local.functions
  name     = "lpa-store/lambda/api-${each.key}"
  provider = aws.management
}

resource "aws_iam_role_policy" "lambda" {
  for_each = local.functions
  name     = "LambdaAllowDynamoDB"
  role     = module.lambda[each.key].iam_role_id
  policy   = data.aws_iam_policy_document.lambda_access_ddb.json
  provider = aws.region
}

data "aws_iam_policy_document" "lambda_access_ddb" {
  statement {
    sid       = "allowDynamoDB"
    effect    = "Allow"
    resources = [var.dynamodb_arn]
    actions = [
      "dynamodb:PutItem",
      "dynamodb:GetItem",
    ]
  }
}
