output "function_url" {
  description = "Public URL of Lambda function"
  value       = aws_lambda_function_url.main.function_url
}

output "iam_role_id" {
  description = "ID of IAM role created for lambda"
  value       = aws_iam_role.lambda.id
}

output "invoke_arn" {
  description = "value"
  value       = aws_lambda_function.main.invoke_arn
}
