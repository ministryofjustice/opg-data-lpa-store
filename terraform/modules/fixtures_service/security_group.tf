resource "aws_security_group" "ecs" {
  name   = "fixtures-${var.environment_name}"
  vpc_id = var.vpc_id

  provider = aws.region
}

resource "aws_security_group_rule" "ecs_to_vpc_endpoint" {
  type                     = "egress"
  protocol                 = "tcp"
  from_port                = 443
  to_port                  = 443
  security_group_id        = aws_security_group.ecs.id
  source_security_group_id = data.aws_security_group.vpc_endpoints_application.id

  provider = aws.region
}

data "aws_security_group" "vpc_endpoints_application" {
  vpc_id = var.vpc_id
  name   = "vpc-endpoint-access-application-subnets-${data.aws_region.current.name}"

  provider = aws.region
}
