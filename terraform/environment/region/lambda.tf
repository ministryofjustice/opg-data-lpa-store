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

  account_name          = var.account_name
  environment_name      = var.environment_name
  lambda_name           = each.key
  ecr_image_uri         = "${data.aws_ecr_repository.lambda[each.key].repository_url}:${var.app_version}"
  event_bus_arn         = var.event_bus.arn
  cloudwatch_kms_key_id = aws_kms_key.cloudwatch.arn
  subnet_ids            = data.aws_subnets.application.ids
  vpc_id                = data.aws_vpc.main.id

  environment_variables = {
    DDB_TABLE_NAME_DEEDS    = var.dynamodb_name
    DDB_TABLE_NAME_CHANGES  = var.dynamodb_name_changes
    EVENT_BUS_NAME          = var.event_bus.name
    S3_BUCKET_NAME_ORIGINAL = var.lpa_store_static_bucket.bucket
    JWT_SECRET_KEY_ID       = "${var.account_name}/jwt-key"
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
