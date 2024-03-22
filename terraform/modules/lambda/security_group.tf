resource "aws_security_group" "lambda" {
  name   = "${var.lambda_name}-${var.environment_name}"
  vpc_id = var.vpc_id
}

resource "aws_security_group_rule" "lambda_to_vpc_endpoint" {
  type                     = "egress"
  protocol                 = "tcp"
  from_port                = 443
  to_port                  = 443
  security_group_id        = aws_security_group.lambda.id
  source_security_group_id = data.aws_security_group.vpc_endpoints_application.id
}

data "aws_security_group" "vpc_endpoints_application" {
  vpc_id = var.vpc_id
  name   = "vpc-endpoint-access-application-subnets-${data.aws_region.current.name}"
}
