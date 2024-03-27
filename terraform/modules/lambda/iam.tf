resource "aws_iam_role" "lambda" {
  name               = "lambda-${var.lambda_name}-${var.environment_name}-${data.aws_region.current.name}"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json

  lifecycle {
    create_before_destroy = true
  }
}

data "aws_iam_policy_document" "lambda_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "aws_xray_write_only_access" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/AWSXrayWriteOnlyAccess"
}

resource "aws_iam_role_policy_attachment" "vpc_access_execution_role" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

resource "aws_iam_role_policy" "lambda" {
  name   = "LambdaRolePermissions"
  role   = aws_iam_role.lambda.id
  policy = data.aws_iam_policy_document.lambda.json
}

data "aws_iam_policy_document" "lambda" {
  statement {
    sid       = "allowLogging"
    effect    = "Allow"
    resources = [aws_cloudwatch_log_group.lambda.arn]
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DescribeLogStreams"
    ]
  }

  statement {
    sid       = "allowPutEvents"
    effect    = "Allow"
    resources = [var.event_bus_arn]
    actions = [
      "events:PutEvents"
    ]
  }

  statement {
    sid       = "allowReadJwtSecret"
    effect    = "Allow"
    resources = [aws_secretsmanager_secret.jwt_secret_key.arn]
    actions = [
      "secretsmanager:GetSecretValue"
    ]
  }
}

data "aws_secretsmanager_secret" "jwt_secret_key" {
  name = "${var.account_name}/jwt-key"
}

resource "aws_lambda_permission" "allow_lambda_execution_operator" {
  statement_id  = "AllowExecutionOperator"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.main.function_name
  principal     = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/operator"
}
