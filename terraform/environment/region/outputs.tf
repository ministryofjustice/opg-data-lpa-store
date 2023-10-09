output "lambda_url" {
  description = "Public URL of 'create' Lambda function"
  value       = module.lambda["create"].function_url
}
