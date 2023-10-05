module "eu_west_1" {
  source = "./region"

  environment_name = local.environment_name
  app_version      = local.app_version

  providers = {
    aws.region = aws.eu_west_1
  }
}

module "eu_west_2" {
  source = "./region"

  environment_name = local.environment_name
  app_version      = local.app_version

  providers = {
    aws.region = aws.eu_west_2
  }
}
