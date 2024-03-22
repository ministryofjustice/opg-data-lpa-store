resource "aws_security_group" "ecs" {
  name   = "fixtures-${var.environment_name}"
  vpc_id = var.vpc_id
}
