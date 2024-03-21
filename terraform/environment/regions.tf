module "eu_west_1" {
  source = "./region"

  allowed_arns                    = local.environment.allowed_arns
  app_version                     = var.app_version
  dns_weighting                   = 100
  dynamodb_arn                    = aws_dynamodb_table.deeds_table.arn
  dynamodb_arn_changes            = aws_dynamodb_table.changes_table.arn
  dynamodb_name                   = aws_dynamodb_table.deeds_table.name
  dynamodb_name_changes           = aws_dynamodb_table.changes_table.name
  environment_name                = local.environment_name
  event_bus                       = aws_cloudwatch_event_bus.main
  has_fixtures                    = local.environment.has_fixtures
  lpa_store_static_bucket         = module.s3_lpa_store_static_eu_west_1.bucket
  lpa_store_static_bucket_kms_key = module.s3_lpa_store_static_eu_west_1.encryption_kms_key

  providers = {
    aws.global     = aws.global
    aws.region     = aws.eu_west_1
    aws.management = aws.management_eu_west_1
  }
}

module "eu_west_2" {
  source = "./region"

  allowed_arns                    = local.environment.allowed_arns
  app_version                     = var.app_version
  dns_weighting                   = 0
  dynamodb_arn                    = aws_dynamodb_table_replica.deeds_table.arn
  dynamodb_arn_changes            = aws_dynamodb_table_replica.changes_table.arn
  dynamodb_name                   = aws_dynamodb_table.deeds_table.name
  dynamodb_name_changes           = aws_dynamodb_table.changes_table.name
  environment_name                = local.environment_name
  event_bus                       = aws_cloudwatch_event_bus.main
  has_fixtures                    = false
  lpa_store_static_bucket         = module.s3_lpa_store_static_eu_west_2.bucket
  lpa_store_static_bucket_kms_key = module.s3_lpa_store_static_eu_west_2.encryption_kms_key

  providers = {
    aws.global     = aws.global
    aws.region     = aws.eu_west_2
    aws.management = aws.management_eu_west_2
  }
}
