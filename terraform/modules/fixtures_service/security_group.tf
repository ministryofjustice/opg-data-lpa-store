resource "aws_security_group" "ecs" {
  name   = "fixtures-${var.environment_name}"
  vpc_id = var.vpc_id

  provider = aws.region
}

resource "aws_security_group_rule" "alb_ingress" {
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 80
  to_port                  = 8080
  source_security_group_id = aws_security_group.loadbalancer_gov_wifi.id
  security_group_id        = aws_security_group.ecs.id
  description              = "Inbound from the ALB"

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

# Allow public web egress so we can access LPA Store APIs
resource "aws_security_group_rule" "ecs_to_public_web" {
  type              = "egress"
  protocol          = "tcp"
  from_port         = 443
  to_port           = 443
  security_group_id = aws_security_group.ecs.id
  cidr_blocks       = ["0.0.0.0/0"]

  provider = aws.region
}

resource "aws_security_group_rule" "ecs_to_vpc_gateways" {
  type              = "egress"
  protocol          = "tcp"
  from_port         = 443
  to_port           = 443
  security_group_id = aws_security_group.ecs.id
  prefix_list_ids   = [data.aws_prefix_list.s3.id]

  provider = aws.region
}

data "aws_prefix_list" "s3" {
  name = "com.amazonaws.${data.aws_region.current.name}.s3"

  provider = aws.region
}
