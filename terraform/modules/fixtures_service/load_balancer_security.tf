module "allow_list" {
  source = "git@github.com:ministryofjustice/opg-terraform-aws-moj-ip-allow-list.git?ref=v3.0.1"
}

locals {
  gov_wifi_allow_lists = {
    embankment_house = {
      label = "Embankment House"
      list  = module.allow_list.gov_wifi_embankment_house
    }
    centre_city = {
      label = "Centre City"
      list  = module.allow_list.gov_wifi_centre_city
    }
    victoria_square_house = {
      label = "Victoria Square House"
      list  = module.allow_list.gov_wifi_victoria_square_house
    }
  }
  moj_sites_access        = module.allow_list.moj_sites
  palo_alto_prisma_access = module.allow_list.palo_alto_prisma_access
}

resource "aws_security_group" "loadbalancer_gov_wifi" {
  name                   = "loadbalancer-gov-wifi-${var.environment_name}"
  description            = "Allow inbound traffic from Gov-Wifi"
  revoke_rules_on_delete = true
  vpc_id                 = var.vpc_id

  lifecycle {
    create_before_destroy = true
  }

  tags = { "Name" = "loadbalancer-gov-wifi-${var.environment_name}" }

  provider = aws.region
}

resource "aws_security_group_rule" "loadbalancer_ingress_http_gov_wifi" {
  for_each          = local.gov_wifi_allow_lists
  type              = "ingress"
  from_port         = 80
  to_port           = 80
  protocol          = "tcp"
  cidr_blocks       = each.value.list
  security_group_id = aws_security_group.loadbalancer_gov_wifi.id
  description       = "Loadbalancer HTTP inbound from ${each.value.label}"

  provider = aws.region
}

resource "aws_security_group_rule" "loadbalancer_ingress_https_gov_wifi" {
  for_each          = local.gov_wifi_allow_lists
  type              = "ingress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = each.value.list
  security_group_id = aws_security_group.loadbalancer_gov_wifi.id
  description       = "Loadbalancer HTTPS inbound from ${each.value.label}"

  provider = aws.region
}

resource "aws_security_group" "loadbalancer_moj_sites_access" {
  name                   = "loadbalancer-moj-sites-access-access-${var.environment_name}"
  description            = "Allow inbound traffic from MOJ Sites"
  revoke_rules_on_delete = true
  vpc_id                 = var.vpc_id

  lifecycle {
    create_before_destroy = true
  }

  tags = { "Name" = "loadbalancer-moj-sites-access-access-${var.environment_name}" }

  provider = aws.region
}

resource "aws_security_group_rule" "loadbalancer_ingress_http_moj_sites_access" {
  type              = "ingress"
  from_port         = 80
  to_port           = 80
  protocol          = "tcp"
  cidr_blocks       = local.moj_sites_access
  security_group_id = aws_security_group.loadbalancer_moj_sites_access.id
  description       = "Loadbalancer HTTP inbound from MOJ Sites"

  provider = aws.region
}

resource "aws_security_group_rule" "loadbalancer_ingress_https_moj_sites_access" {
  type              = "ingress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = local.moj_sites_access
  security_group_id = aws_security_group.loadbalancer_moj_sites_access.id
  description       = "Loadbalancer HTTPS inbound from MOJ Sites"

  provider = aws.region
}

resource "aws_security_group_rule" "loadbalancer_ingress_http_palo_alto_prisma_access" {
  type              = "ingress"
  from_port         = 80
  to_port           = 80
  protocol          = "tcp"
  cidr_blocks       = local.palo_alto_prisma_access
  security_group_id = aws_security_group.loadbalancer_moj_sites_access.id
  description       = "Loadbalancer HTTP inbound from Palo Alto Prisma"

  provider = aws.region
}

resource "aws_security_group_rule" "loadbalancer_ingress_https_palo_alto_prisma_access" {
  type              = "ingress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = local.palo_alto_prisma_access
  security_group_id = aws_security_group.loadbalancer_moj_sites_access.id
  description       = "Loadbalancer HTTPS inbound from Palo Alto Prisma"

  provider = aws.region
}

resource "aws_security_group" "loadbalancer_ingress_route53" {
  name                   = "loadbalancer-route53-${var.environment_name}"
  description            = "Allow inbound traffic from Route53"
  revoke_rules_on_delete = true
  vpc_id                 = var.vpc_id

  lifecycle {
    create_before_destroy = true
  }

  tags = { "Name" = "lb-route53-${var.environment_name}" }

  provider = aws.region
}

data "aws_ip_ranges" "route53_healthchecks" {
  services = ["route53_healthchecks"]

  provider = aws.region
}

resource "aws_security_group_rule" "loadbalancer_ingress_route53_healthchecks" {
  type              = "ingress"
  protocol          = "tcp"
  from_port         = "443"
  to_port           = "443"
  cidr_blocks       = data.aws_ip_ranges.route53_healthchecks.cidr_blocks
  security_group_id = aws_security_group.loadbalancer_ingress_route53.id
  description       = "Loadbalancer ingresss from Route53 healthchecks"

  provider = aws.region
}
