data "aws_secretsmanager_secret" "jwt_secret_key" {
  name = "${var.account_name}/jwt-key"

  provider = aws.region
}
