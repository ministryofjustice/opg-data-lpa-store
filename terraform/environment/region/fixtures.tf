module "fixtures" {
  count  = var.has_fixtures ? 1 : 0
  source = "../../modules/fixtures_service"

  environment_name      = var.environment_name
  cloudwatch_kms_key_id = aws_kms_key.cloudwatch.arn
  service_url           = local.domain_name

  providers = {
    aws.global     = aws.global
    aws.management = aws.management
    aws.region     = aws.region
  }
}
