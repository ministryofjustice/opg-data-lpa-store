resource "aws_secretsmanager_secret" "jwt_key" {
  name = "${var.account_name}/jwt-key"

  provider = aws.region
}
