resource "aws_lb" "load_balancer" {
  name                       = var.environment_name
  internal                   = false
  load_balancer_type         = "application"
  subnets                    = var.public_subnet_ids
  drop_invalid_header_fields = true
  enable_deletion_protection = false

  security_groups = flatten([
    aws_security_group.loadbalancer_gov_wifi.id,
    aws_security_group.loadbalancer_moj_sites_access.id,
    aws_security_group.loadbalancer_ingress_route53.id,
  ])

  tags = { "Name" = "lb-${var.environment_name}" }

  provider = aws.region
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.load_balancer.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }

  provider = aws.region
}

resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.load_balancer.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-FS-1-2-Res-2020-10"
  certificate_arn   = data.aws_acm_certificate.root.arn

  default_action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "LPA Store ${var.environment_name} ALB"
      status_code  = "200"
    }
  }

  provider = aws.region
}

resource "aws_lb_listener_rule" "fixtures" {
  listener_arn = aws_lb_listener.https.arn
  priority     = 10

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.fixtures.arn
  }

  condition {
    host_header {
      values = ["fixtures-${var.environment_name}.*"]
    }
  }

  provider = aws.region
}

resource "aws_lb_target_group" "fixtures" {
  name                 = "fixtures-${var.environment_name}-http"
  port                 = 80
  protocol             = "HTTP"
  target_type          = "ip"
  vpc_id               = var.vpc_id
  deregistration_delay = 0
  depends_on           = [aws_lb.load_balancer]

  # health_check {
  #   protocol            = "HTTP"
  #   path                = "/health-check"
  #   interval            = 15
  #   timeout             = 10
  #   healthy_threshold   = 2
  #   unhealthy_threshold = 5
  #   matcher             = "200"
  # }

  lifecycle {
    create_before_destroy = true
  }

  provider = aws.region
}
