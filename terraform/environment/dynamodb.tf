resource "aws_dynamodb_table" "deeds_table" {
  name                        = "deeds-${local.environment_name}"
  billing_mode                = "PAY_PER_REQUEST"
  deletion_protection_enabled = local.environment.is_production
  stream_enabled              = true
  stream_view_type            = "NEW_AND_OLD_IMAGES"
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

  provider = aws.eu_west_1
}

resource "aws_dynamodb_table_replica" "deeds_table" {
  global_table_arn       = aws_dynamodb_table.deeds_table.arn
  point_in_time_recovery = true
  provider               = aws.eu_west_2
}

resource "aws_dynamodb_table" "changes_table" {
  name                        = "changes-${local.environment_name}"
  billing_mode                = "PAY_PER_REQUEST"
  deletion_protection_enabled = local.environment.is_production
  stream_enabled              = true
  stream_view_type            = "NEW_AND_OLD_IMAGES"
  hash_key                    = "uid"

  server_side_encryption {
    enabled = true
  }

  attribute {
    name = "uid"
    type = "S"
  }

  range_key = "time"

  attribute {
    name = "time"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }

  lifecycle {
    ignore_changes = [replica]
  }

  provider = aws.eu_west_1
}

resource "aws_dynamodb_table_replica" "changes_table" {
  global_table_arn       = aws_dynamodb_table.changes_table.arn
  point_in_time_recovery = true
  provider               = aws.eu_west_2
}
