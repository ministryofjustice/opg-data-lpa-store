output "base_url" {
  value = aws_api_gateway_stage.current.invoke_url
}
