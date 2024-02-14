resource "aws_cloudwatch_event_bus" "main" {
  name = "${local.environment_name}-main"

  provider = aws.eu_west_1
}

resource "aws_cloudwatch_event_archive" "main" {
  name             = "${local.environment_name}-main"
  event_source_arn = aws_cloudwatch_event_bus.main.arn

  provider = aws.eu_west_1
}

resource "aws_cloudwatch_event_rule" "source_events_to_sirius" {
  event_bus_name = aws_cloudwatch_event_bus.main.name
  name           = "${local.environment_name}-source-events-to-sirius"
  state          = "ENABLED"
  event_pattern = jsonencode({
    source = ["opg.poas.lpastore"]
  })

  provider = aws.eu_west_1
}

resource "aws_kms_key" "event_dlq" {
  description         = "KMS Key for ${local.environment_name} event dead-letter queue"
  enable_key_rotation = true
  policy              = data.aws_iam_policy_document.event_dlq_kms_key.json
  tags                = { "Name" = "event-dlq-kms-key-${local.environment_name}" }

  provider = aws.eu_west_1
}

resource "aws_kms_alias" "event_dlq" {
  name          = "alias/${local.environment_name}-event-dlq"
  target_key_id = aws_kms_key.event_dlq.id

  depends_on = [aws_kms_key.event_dlq]
  provider   = aws.eu_west_1
}

resource "aws_sqs_queue" "event_dlq" {
  name              = "${local.environment_name}-events-dlq"
  kms_master_key_id = aws_kms_alias.event_dlq.name

  provider = aws.eu_west_1
}

resource "aws_sqs_queue_policy" "deadletter_queue" {
  queue_url = aws_sqs_queue.event_dlq.id
  policy    = data.aws_iam_policy_document.deadletter_queue.json

  provider = aws.eu_west_1
}

data "aws_iam_policy_document" "event_dlq_kms_key" {
  statement {
    sid       = "Enable IAM User Permissions"
    effect    = "Allow"
    actions   = ["kms:*"]
    resources = ["*"]

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${local.environment.account_id}:root"]
    }
  }

  statement {
    sid    = "Allow EventBridge to use key"
    effect = "Allow"
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey"
    ]
    resources = ["*"]

    principals {
      type        = "Service"
      identifiers = ["events.amazonaws.com"]
    }
  }

  provider = aws.eu_west_1
}

data "aws_iam_policy_document" "deadletter_queue" {
  statement {
    sid       = "allowEB-${aws_cloudwatch_event_bus.main.name}"
    effect    = "Allow"
    resources = [aws_sqs_queue.event_dlq.arn]
    actions = [
      "sqs:SendMessage",
    ]
    principals {
      type        = "Service"
      identifiers = ["events.amazonaws.com"]
    }
    condition {
      test     = "ArnEquals"
      variable = "aws:SourceArn"
      values   = [aws_cloudwatch_event_rule.source_events_to_sirius.arn]
    }
  }

  provider = aws.eu_west_1
}

module "eventbridge_cross_account_target" {
  for_each              = toset(local.environment.target_event_buses)
  source                = "../modules/eventbridge_cross_account_target"
  name_suffix           = "${local.environment_name}-sirius"
  rule                  = aws_cloudwatch_event_rule.source_events_to_sirius
  dead_letter_queue_arn = aws_sqs_queue.event_dlq.arn
  target_event_bus_arn  = each.value

  providers = {
    aws = aws.eu_west_1
  }
}
