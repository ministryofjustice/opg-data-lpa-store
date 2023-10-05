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
}

resource "aws_lambda_function_url" "main" {
  function_name      = aws_lambda_function.main.function_name
  authorization_type = "NONE"
}
