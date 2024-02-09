output "base_url" {
  value = "https://${local.domain_name}"
}

output "lambda_iam_roles" {
  value = [
    for lambda in module.lambda : lambda.iam_role
  ]
}
