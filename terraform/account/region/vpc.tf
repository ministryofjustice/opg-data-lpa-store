module "vpc" {
  source                         = "github.com/ministryofjustice/opg-terraform-aws-network?ref=v1.3.3"
  cidr                           = var.vpc_cidr
  enable_dns_hostnames           = true
  enable_dns_support             = true
  default_security_group_ingress = []
  default_security_group_egress  = []
  providers = {
    aws = aws.region
  }
}
