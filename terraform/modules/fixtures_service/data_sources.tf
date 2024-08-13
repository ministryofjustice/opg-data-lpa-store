data "aws_secretsmanager_secret" "jwt_secret_key" {
  name     = "${data.aws_default_tags.default.tags.application}/${data.aws_default_tags.default.tags.account}/jwt-key"
  provider = aws.management
}

data "aws_kms_alias" "jwt_key" {
  name     = "alias/${data.aws_default_tags.default.tags.application}/${data.aws_default_tags.default.tags.account}/jwt-key"
  provider = aws.management
}

data "aws_default_tags" "default" {
  provider = aws.region
}
