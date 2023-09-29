resource "aws_dynamodb_table" "deeds_table" {
  name                        = "${local.environment_name}-deeds"
  billing_mode                = "PAY_PER_REQUEST"
  deletion_protection_enabled = local.environment.is_production
  stream_enabled              = false
  hash_key                    = "uid"

  server_side_encryption {
    enabled = true
  }

  attribute {
    name = "uid"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }

  lifecycle {
    ignore_changes = [replica]
  }
}

resource "aws_dynamodb_table_replica" "deeds_table" {
  global_table_arn       = aws_dynamodb_table.deeds_table.arn
  point_in_time_recovery = true
  provider               = aws.eu-west-2
}
