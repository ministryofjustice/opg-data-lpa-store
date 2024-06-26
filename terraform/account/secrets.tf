resource "aws_secretsmanager_secret" "jwt_key" {
  name        = "${data.aws_default_tags.default.tags.application}/${data.aws_default_tags.default.tags.account}/jwt-key"
  description = "JWT key for ${data.aws_default_tags.default.tags.application} in ${data.aws_default_tags.default.tags.account}, for use with Make and Register, and Use a LPA"
  replica {
    region = data.aws_region.eu_west_2.name
  }
  provider = aws.management_eu_west_1
}
