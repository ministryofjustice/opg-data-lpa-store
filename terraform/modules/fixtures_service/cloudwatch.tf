resource "aws_cloudwatch_log_group" "fixtures" {
  name              = "/ecs/fixtures-${var.environment_name}"
  kms_key_id        = var.cloudwatch_kms_key_id
  retention_in_days = 400

  provider = aws.region
}
