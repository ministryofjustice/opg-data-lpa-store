resource "aws_cloudwatch_event_target" "cross_account_event_bus" {
  target_id      = "cross-put-${var.name_suffix}"
  event_bus_name = var.rule.event_bus_name
  rule           = var.rule.name
  arn            = var.target_event_bus_arn
  role_arn       = aws_iam_role.cross_account_put.arn

  dead_letter_config {
    arn = var.dead_letter_queue_arn
  }
}

resource "aws_iam_role" "cross_account_put" {
  name               = "cross-put-${var.name_suffix}"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy" "cross_account_put" {
  name   = "cross-put-${var.name_suffix}"
  policy = data.aws_iam_policy_document.cross_account_put_access.json
  role   = aws_iam_role.cross_account_put.id
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["events.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

data "aws_iam_policy_document" "cross_account_put_access" {
  statement {
    sid    = "CrossAccountPutAccess"
    effect = "Allow"
    actions = [
      "events:PutEvents",
    ]
    resources = [
      var.target_event_bus_arn
    ]
  }
}
