resource "aws_vpc" "main" {
  cidr_block           = var.cidr
  instance_tenancy     = var.instance_tenancy
  enable_dns_hostnames = var.enable_dns_hostnames
  enable_dns_support   = var.enable_dns_support
  tags                 = { Name = "${local.name-prefix}-vpc" }
}

resource "aws_vpc_dhcp_options" "dns_resolver" {
  domain_name         = var.dhcp_options_domain_name
  domain_name_servers = var.dhcp_options_domain_name_servers
}

resource "aws_vpc_dhcp_options_association" "dns_resolver" {
  vpc_id          = aws_vpc.main.id
  dhcp_options_id = aws_vpc_dhcp_options.dns_resolver.id
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.main.id

  tags = { Name = "${local.name-prefix}-internet-gateway" }
}

resource "aws_eip" "nat" {
  count  = 3
  domain = "vpc"
  tags   = { Name = "${local.name-prefix}-eip-nat-gateway" }
}

resource "aws_nat_gateway" "gw" {
  count         = 3
  allocation_id = aws_eip.nat[count.index].id
  subnet_id     = aws_subnet.public[count.index].id
  tags          = { Name = "${local.name-prefix}-nat-gateway-${data.aws_availability_zones.all.names[count.index]}" }
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = aws_vpc.main.default_network_acl_id
  subnet_ids = concat(
    aws_subnet.public[*].id,
    aws_subnet.application[*].id,
    aws_subnet.data[*].id
  )

  egress {
    action     = "allow"
    cidr_block = "0.0.0.0/0"
    from_port  = 0
    icmp_code  = 0
    icmp_type  = 0
    protocol   = "-1" #tfsec:ignore:aws-ec2-no-excessive-port-access
    rule_no    = 10
    to_port    = 0
  }
  ingress {
    action     = "deny"
    cidr_block = "0.0.0.0/0" #tfsec:ignore:aws-ec2-no-public-ingress-acl
    from_port  = 22
    icmp_code  = 0
    icmp_type  = 0
    protocol   = "6"
    rule_no    = 120
    to_port    = 22
  }
  ingress {
    action     = "deny"
    cidr_block = "0.0.0.0/0" #tfsec:ignore:aws-ec2-no-public-ingress-acl
    from_port  = 3389
    icmp_code  = 0
    icmp_type  = 0
    protocol   = "6"
    rule_no    = 130
    to_port    = 3389
  }
  ingress {
    action          = "deny"
    ipv6_cidr_block = "::/0"
    from_port       = 22
    icmp_code       = 0
    icmp_type       = 0
    protocol        = "6"
    rule_no         = 125
    to_port         = 22
  }
  ingress {
    action          = "deny"
    ipv6_cidr_block = "::/0"
    from_port       = 3389
    icmp_code       = 0
    icmp_type       = 0
    protocol        = "6"
    rule_no         = 135
    to_port         = 3389
  }
  ingress {
    action     = "allow"
    cidr_block = "0.0.0.0/0"
    from_port  = 0
    icmp_code  = 0
    icmp_type  = 0
    protocol   = "-1" #tfsec:ignore:aws-ec2-no-excessive-port-access
    rule_no    = 160
    to_port    = 0
  }
}
