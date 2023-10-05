locals {
  functions = [
    "create",
    "get",
    "update",
  ]
}

module "lambda" {
  for_each = local.functions
  source   = "../../modules/lambda"

  environment_name = var.environment_name
  lambda_name      = each.value
  ecr_image_uri    = "${aws_ecr_repository.lambda[each.value].repository_url}:${var.app_version}"
}

data "aws_ecr_repository" "lambda" {
  name     = "lpa-store/lambda/api-${each.value}"
  provider = aws.management
}
