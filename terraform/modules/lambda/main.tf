resource "aws_cloudwatch_log_group" "lambda" {
  name       = "/aws/lambda/${var.lambda_name}-${var.environment_name}"
  kms_key_id = var.cloudwatch_kms_key_id
}

resource "aws_lambda_function" "main" {
  function_name = "${var.lambda_name}-${var.environment_name}"
  image_uri     = var.ecr_image_uri
  package_type  = "Image"
  role          = aws_iam_role.lambda.arn
  timeout       = 5
  depends_on    = [aws_cloudwatch_log_group.lambda]

  tracing_config {
    mode = "Active"
  }

  vpc_config {
    subnet_ids         = var.subnet_ids
    security_group_ids = [aws_security_group.lambda.id]
  }

  dynamic "environment" {
    for_each = length(keys(var.environment_variables)) == 0 ? [] : [true]
    content {
      variables = var.environment_variables
    }
  }
}
