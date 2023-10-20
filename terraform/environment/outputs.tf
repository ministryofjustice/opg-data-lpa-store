output "lambda_url" {
  description = "Public URL of 'create' Lambda function"
  value       = module.eu_west_1.lambda_url
}

output "base_url" {
  description = "Base URL of API"
  value       = module.eu_west_1.base_url
}
