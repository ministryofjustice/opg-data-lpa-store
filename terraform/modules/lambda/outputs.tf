output "iam_role_id" {
  description = "ID of IAM role created for lambda"
  value       = aws_iam_role.lambda.id
}

output "invoke_arn" {
  description = "Invoke ARN of Lambda function"
  value       = aws_lambda_function.main.invoke_arn
}

output "function_name" {
  description = "Name of Lambda function"
  value       = aws_lambda_function.main.function_name
}
