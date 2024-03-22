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

resource "aws_security_group" "vpc_endpoints_application" {
  name   = "vpc-endpoint-access-application-subnets-${data.aws_region.current.name}"
  vpc_id = module.vpc.vpc.id
}

resource "aws_security_group_rule" "vpc_endpoints_application_subnet_ingress" {
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.vpc_endpoints_application.id
  type              = "ingress"
  cidr_blocks       = module.vpc.application_subnets[*].cidr_block
  description       = "Allow Services in Private Subnets of ${data.aws_region.current.name} to connect to VPC Interface Endpoints"
}

locals {
  interface_endpoint = toset([
    "dynamodb",
    "ecr.api",
    "ecr.dkr",
    "execute-api",
    "logs",
    "s3",
    "secretsmanager",
    "xray",
  ])
}

resource "aws_vpc_endpoint" "application" {
  for_each = local.interface_endpoint

  vpc_id              = module.vpc.vpc.id
  service_name        = "com.amazonaws.${data.aws_region.current.name}.${each.value}"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = true
  security_group_ids  = aws_security_group.vpc_endpoints_application[*].id
  subnet_ids          = module.vpc.application_subnets[*].id
}
