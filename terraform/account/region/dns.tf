locals {
  domain_name = var.is_production ? data.aws_route53_zone.service.name : "*.${data.aws_route53_zone.service.name}"
}

data "aws_route53_zone" "service" {
  name     = "lpa-store.api.opg.service.justice.gov.uk"
  provider = aws.management
}

resource "aws_acm_certificate" "environment" {
  domain_name       = local.domain_name
  validation_method = "DNS"
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
