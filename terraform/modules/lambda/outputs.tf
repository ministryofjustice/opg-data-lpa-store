output "function_url" {
  description = "Public URL of Lambda function"
  value       = aws_lambda_function_url.main.function_url
}
