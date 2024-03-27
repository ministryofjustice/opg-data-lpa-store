module "fixtures" {
  count  = var.has_fixtures ? 1 : 0
  source = "../../modules/fixtures_service"

  account_name           = var.account_name
  application_subnet_ids = data.aws_subnets.application.ids
  cloudwatch_kms_key_id  = aws_kms_key.cloudwatch.arn
  ecr_image_uri          = "${data.aws_ecr_repository.fixtures[0].repository_url}:${var.app_version}"
  environment_name       = var.environment_name
  public_subnet_ids      = data.aws_subnets.public.ids
  service_url            = local.domain_name
  vpc_id                 = data.aws_vpc.main.id

  providers = {
    aws.global     = aws.global
    aws.management = aws.management
    aws.region     = aws.region
  }
}

data "aws_ecr_repository" "fixtures" {
  count    = var.has_fixtures ? 1 : 0
  name     = "lpa-store/fixtures"
  provider = aws.management
}
