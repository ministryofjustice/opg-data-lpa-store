resource "aws_cloudwatch_log_group" "lambda" {
  name       = "/aws/lambda/${var.environment_name}-${var.lambda_name}"
  kms_key_id = var.kms_key
  provider   = aws.region
}

resource "aws_lambda_function" "main" {
  function_name = "${var.lambda_name}-${var.environment_name}"
  image_uri     = var.ecr_arn
  package_type  = "Image"
  role          = aws_iam_role.lambda.arn
  timeout       = 5
  depends_on    = [aws_cloudwatch_log_group.lambda]

  tracing_config {
    mode = "Active"
  }

  dynamic "environment" {
    for_each = length(keys(var.environment_variables)) == 0 ? [] : [true]
    content {
      variables = var.environment_variables
    }
  }
  provider = aws.region
}

resource "aws_lambda_function_url" "main" {
  function_name      = aws_lambda_function.main.function_name
  authorization_type = "NONE"
}
