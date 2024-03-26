module "eu_west_1" {
  source = "./region"

  account_name  = local.account.account_name
  is_production = local.account.is_production
  vpc_cidr      = "10.162.0.0/16"

  providers = {
    aws.region     = aws.eu_west_1
    aws.management = aws.management_eu_west_1
  }
}

module "eu_west_2" {
  source = "./region"

  account_name  = local.account.account_name
  is_production = local.account.is_production
  vpc_cidr      = "10.163.0.0/16"

  providers = {
    aws.region     = aws.eu_west_2
    aws.management = aws.management_eu_west_2
  }
}
