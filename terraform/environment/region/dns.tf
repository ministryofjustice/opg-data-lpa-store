locals {
  a_record = terraform.workspace == "production" ? data.aws_route53_zone.service.name : var.environment_name
}

data "aws_route53_zone" "service" {
  name     = "lpa-store.api.opg.service.justice.gov.uk"
  provider = aws.management
}

resource "aws_acm_certificate" "environment" {
  domain_name               = "*.${data.aws_route53_zone.service.name}"
  validation_method         = "DNS"
  subject_alternative_names = [data.aws_route53_zone.service.name]
  lifecycle {
    create_before_destroy = true
  }

  provider = aws.region
}

resource "aws_route53_record" "validation" {
  name            = sort(aws_acm_certificate.environment.domain_validation_options[*].resource_record_name)[0]
  type            = sort(aws_acm_certificate.environment.domain_validation_options[*].resource_record_type)[0]
  zone_id         = data.aws_route53_zone.service.id
  records         = [sort(aws_acm_certificate.environment.domain_validation_options[*].resource_record_value)[0]]
  ttl             = 60
  allow_overwrite = true
  provider        = aws.management
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
  domain_name              = terraform.workspace == "production" ? data.aws_route53_zone.service.name : "${local.a_record}.${data.aws_route53_zone.service.name}"
  regional_certificate_arn = aws_acm_certificate.environment.arn
  security_policy          = "TLS_1_2"

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  provider = aws.region
}
