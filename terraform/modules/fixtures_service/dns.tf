data "aws_acm_certificate" "root" {
  domain   = terraform.workspace == "production" ? data.aws_route53_zone.service.name : "*.${data.aws_route53_zone.service.name}"
  provider = aws.region
}

data "aws_route53_zone" "service" {
  name     = "lpa-store.api.opg.service.justice.gov.uk"
  provider = aws.management
}

resource "aws_route53_record" "fixtures" {
  zone_id        = data.aws_route53_zone.service.zone_id
  name           = "fixtures-${var.environment_name}"
  type           = "A"
  set_identifier = data.aws_region.current.name

  weighted_routing_policy {
    weight = 100
  }

  alias {
    evaluate_target_health = false
    name                   = aws_lb.load_balancer.dns_name
    zone_id                = aws_lb.load_balancer.zone_id
  }

  provider = aws.management
}
