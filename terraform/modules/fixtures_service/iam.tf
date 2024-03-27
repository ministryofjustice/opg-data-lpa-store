resource "aws_iam_role" "task_role" {
  name_prefix        = "fixtures-task-role-${var.environment_name}-"
  assume_role_policy = data.aws_iam_policy_document.ecs_task_role_assume_policy.json

  provider = aws.global
}

data "aws_iam_policy_document" "ecs_task_role_assume_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      identifiers = ["ecs-tasks.amazonaws.com"]
      type        = "Service"
    }
  }

  provider = aws.region
}

resource "aws_iam_role_policy" "task_role" {
  name   = "fixtures-task-role-${var.environment_name}-${data.aws_region.current.name}"
  role   = aws_iam_role.task_role.id
  policy = data.aws_iam_policy_document.task_role.json

  provider = aws.region
}

data "aws_iam_policy_document" "task_role" {
  statement {
    sid    = "AllowInvokeOnLpaStoreRestAPIs"
    effect = "Allow"
    actions = [
      "execute-api:Invoke",
      "execute-api:ManageConnections"
    ]
    resources = ["arn:aws:execute-api:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"]
  }

  provider = aws.region
}


resource "aws_iam_role" "execution_role" {
  name_prefix        = "fixtures-execution-role-${var.environment_name}-"
  assume_role_policy = data.aws_iam_policy_document.execution_assume_role.json

  provider = aws.global
}

data "aws_iam_policy_document" "execution_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      identifiers = ["ecs-tasks.amazonaws.com"]
      type        = "Service"
    }
  }

  provider = aws.region
}

resource "aws_iam_role_policy" "execution_role" {
  name   = "fixtures-execution-role-${var.environment_name}-${data.aws_region.current.name}"
  role   = aws_iam_role.execution_role.id
  policy = data.aws_iam_policy_document.execution_role.json

  provider = aws.region
}

data "aws_iam_policy_document" "execution_role" {
  statement {
    effect    = "Allow"
    resources = ["*"]
    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
    ]
  }
  statement {
    effect = "Allow"
    resources = [
      "${aws_cloudwatch_log_group.fixtures.arn}*",
    ]
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
  }

  statement {
    effect    = "Allow"
    resources = [data.aws_secretsmanager_secret.jwt_secret_key.arn]
    actions = [
      "secretsmanager:GetSecretValue"
    ]
  }

  provider = aws.region
}
