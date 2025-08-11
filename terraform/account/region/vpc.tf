module "vpc" {
  source                         = "github.com/ministryofjustice/opg-terraform-aws-network?ref=v1.6.0"
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

  provider = aws.region
}

resource "aws_security_group_rule" "vpc_endpoints_application_subnet_ingress" {
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.vpc_endpoints_application.id
  type              = "ingress"
  cidr_blocks       = module.vpc.application_subnets[*].cidr_block
  description       = "Allow Services in Application Subnets of ${data.aws_region.current.name} to connect to VPC Interface Endpoints"

  provider = aws.region
}

locals {
  interface_endpoint = toset([
    "ecr.api",
    "ecr.dkr",
    "events",
    "logs",
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
  tags                = { Name = "${each.value}-application-${data.aws_region.current.name}" }

  provider = aws.region
}

data "aws_route_tables" "application" {
  provider = aws.region
  filter {
    name   = "tag:Name"
    values = ["application-route-table"]
  }
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id            = module.vpc.vpc.id
  service_name      = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids   = tolist(data.aws_route_tables.application.ids)
  vpc_endpoint_type = "Gateway"
  policy            = data.aws_iam_policy_document.s3.json
  tags              = { Name = "s3-application-${data.aws_region.current.name}" }

  provider = aws.region
}

resource "aws_vpc_endpoint" "dynamodb" {
  vpc_id            = module.vpc.vpc.id
  service_name      = "com.amazonaws.${data.aws_region.current.name}.dynamodb"
  route_table_ids   = tolist(data.aws_route_tables.application.ids)
  vpc_endpoint_type = "Gateway"
  policy            = data.aws_iam_policy_document.allow_account_access.json
  tags              = { Name = "dynamodb-application-${data.aws_region.current.name}" }

  provider = aws.region
}

data "aws_iam_policy_document" "allow_account_access" {
  provider = aws.region
  statement {
    sid       = "Allow-callers-from-specific-account"
    effect    = "Allow"
    actions   = ["*"]
    resources = ["*"]
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }
}

data "aws_iam_policy_document" "s3" {
  source_policy_documents = [
    data.aws_iam_policy_document.allow_account_access.json,
    data.aws_iam_policy_document.s3_bucket_access.json,
  ]
}

data "aws_iam_policy_document" "s3_bucket_access" {
  statement {
    sid       = "Access-to-specific-bucket-only"
    effect    = "Allow"
    actions   = ["s3:GetObject"]
    resources = ["arn:aws:s3:::prod-${data.aws_region.current.name}-starport-layer-bucket/*"]
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }
}
