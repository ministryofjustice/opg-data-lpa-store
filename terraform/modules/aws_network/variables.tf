data "aws_default_tags" "default_tags" {}

locals {
  name-prefix = "${data.aws_default_tags.default_tags.tags.application}-${data.aws_default_tags.default_tags.tags.environment-name}"
}

variable "cidr" {
  type    = string
  default = "10.0.0.0/16"
}

variable "instance_tenancy" {
  type    = string
  default = "default"
}

variable "enable_dns_hostnames" {
  type    = bool
  default = false
}

variable "enable_dns_support" {
  type    = bool
  default = true
}

variable "dhcp_options_domain_name" {
  type    = string
  default = ""
}

variable "dhcp_options_domain_name_servers" {
  type    = list(string)
  default = ["AmazonProvidedDNS"]
}

variable "map_public_ip_on_launch" {
  type    = bool
  default = false
}

variable "public_subnet_assign_ipv6_address_on_creation" {
  type    = bool
  default = false
}

variable "default_security_group_ingress" {
  type    = list(map(string))
  default = []
}

variable "default_security_group_egress" {
  type    = list(map(string))
  default = []
}

variable "flow_log_log_format" {
  type    = string
  default = null
}

variable "flow_log_traffic_type" {
  description = "The type of traffic to log. Values: ACCEPT, REJECT, ALL."
  type        = string
  default     = "ALL"
}

variable "flow_log_cloudwatch_log_group_retention_in_days" {
  description = "Number of days you want to retain log events."
  type        = number
  default     = null
}

variable "flow_log_cloudwatch_log_group_kms_key_id" {
  type    = string
  default = null
}
