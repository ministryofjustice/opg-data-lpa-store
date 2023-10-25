module "eu_west_1" {
  source = "./region"

  is_production = local.account.is_production

  providers = {
    aws.region     = aws.eu_west_1
    aws.management = aws.management_eu_west_1
  }
}

module "eu_west_2" {
  source = "./region"

  is_production = local.account.is_production

  providers = {
    aws.region     = aws.eu_west_2
    aws.management = aws.management_eu_west_2
  }
}
