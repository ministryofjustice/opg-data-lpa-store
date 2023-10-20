output "lambda_url" {
  description = "Public URL of 'create' Lambda function"
  value       = module.lambda["create"].function_url
}

output "base_url" {
  value = aws_api_gateway_stage.current.invoke_url
}
