locals {
  a_record    = terraform.workspace == "production" ? data.aws_route53_zone.service.name : var.environment_name
  domain_name = terraform.workspace == "production" ? data.aws_route53_zone.service.name : "${local.a_record}.${data.aws_route53_zone.service.name}"
}

data "aws_route53_zone" "service" {
  name     = "lpa-store.api.opg.service.justice.gov.uk"
  provider = aws.management
}

data "aws_acm_certificate" "root" {
  domain   = terraform.workspace == "production" ? data.aws_route53_zone.service.name : "*.${data.aws_route53_zone.service.name}"
  provider = aws.region
}

resource "aws_route53_record" "environment_record" {
  name           = local.a_record
  type           = "A"
  zone_id        = data.aws_route53_zone.service.id
  set_identifier = data.aws_region.current.name

  weighted_routing_policy {
    weight = var.dns_weighting
  }

  alias {
    evaluate_target_health = true
    name                   = aws_api_gateway_domain_name.lpa_store.regional_domain_name
    zone_id                = aws_api_gateway_domain_name.lpa_store.regional_zone_id
  }

  provider = aws.management
}

resource "aws_api_gateway_domain_name" "lpa_store" {
  domain_name              = local.domain_name
  regional_certificate_arn = data.aws_acm_certificate.root.arn
  security_policy          = "TLS_1_2"

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  provider = aws.region
}
