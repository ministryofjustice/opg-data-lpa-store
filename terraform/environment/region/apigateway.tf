locals {
  stage_name = "current"
  template_file = templatefile("../../docs/openapi/openapi-aws.compiled.yaml", {
    lambda_create_invoke_arn = module.lambda["create"].invoke_arn
    lambda_get_invoke_arn    = module.lambda["get"].invoke_arn
    lambda_update_invoke_arn = module.lambda["update"].invoke_arn
    lambda_getall_invoke_arn = module.lambda["getall"].invoke_arn
  })
}

resource "aws_api_gateway_rest_api" "lpa_store" {
  name        = "lpa-store-${var.environment_name}"
  description = "API Gateway for LPA Store - ${var.environment_name}"
  body        = local.template_file

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  provider = aws.region
}


resource "aws_api_gateway_rest_api_policy" "lpa_store" {
  rest_api_id = aws_api_gateway_rest_api.lpa_store.id
  policy      = data.aws_iam_policy_document.lpa_store.json

  provider = aws.region
}


resource "aws_api_gateway_deployment" "lpa_store" {
  rest_api_id = aws_api_gateway_rest_api.lpa_store.id

  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_rest_api.lpa_store.body,
      var.allowed_arns,
      var.allowed_wildcard_arns,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }

  depends_on = [
    aws_api_gateway_rest_api.lpa_store,
    aws_api_gateway_rest_api_policy.lpa_store
  ]

  provider = aws.region
}

resource "aws_api_gateway_stage" "current" {
  deployment_id        = aws_api_gateway_deployment.lpa_store.id
  rest_api_id          = aws_api_gateway_rest_api.lpa_store.id
  stage_name           = local.stage_name
  xray_tracing_enabled = true

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.lpa_store.arn
    format = join("", [
      "{\"requestId\":\"$context.requestId\",",
      "\"ip\":\"$context.identity.sourceIp\",",
      "\"caller\":\"$context.identity.caller\",",
      "\"user\":\"$context.identity.user\",",
      "\"requestTime\":\"$context.requestTime\",",
      "\"httpMethod\":\"$context.httpMethod\"",
      "\"resourcePath\":\"$context.resourcePath\",",
      "\"status\":\"$context.status\",",
      "\"protocol\":\"$context.protocol\",",
      "\"responseLength\":\"$context.responseLength\"}"
    ])
  }

  depends_on = [
    aws_api_gateway_account.api_gateway,
    aws_cloudwatch_log_group.lpa_store
  ]

  provider = aws.region
}

resource "aws_cloudwatch_log_group" "lpa_store" {
  name              = "API-Gateway-Execution-Logs-${aws_api_gateway_rest_api.lpa_store.name}-${local.stage_name}"
  kms_key_id        = aws_kms_key.cloudwatch.arn
  retention_in_days = 400

  provider = aws.region
}

resource "aws_api_gateway_method_settings" "lpa_store_gateway_settings" {
  rest_api_id = aws_api_gateway_rest_api.lpa_store.id
  stage_name  = aws_api_gateway_stage.current.stage_name
  method_path = "*/*"

  settings {
    metrics_enabled = true
    logging_level   = "INFO"
  }

  provider = aws.region
}

data "aws_iam_role" "api_gateway_cloudwatch" {
  name = "api-gateway-cloudwatch-global"

  provider = aws.region
}

resource "aws_api_gateway_account" "api_gateway" {
  cloudwatch_role_arn = data.aws_iam_role.api_gateway_cloudwatch.arn

  provider = aws.region
}


data "aws_iam_policy_document" "lpa_store" {
  policy_id                 = "lpa-store-${var.environment_name}-${data.aws_region.current.name}-resource-policy"
  override_policy_documents = length(var.allowed_wildcard_arns) > 0 ? [data.aws_iam_policy_document.lpa_store_wildcard.json] : null

  statement {
    sid    = "AllowExecutionFromAllowedARNs"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = var.allowed_arns
    }

    actions   = ["execute-api:Invoke"]
    resources = ["*"]
  }

  statement {
    sid    = "AllowHealthCheckExecutionFromAnyone"
    effect = "Allow"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions   = ["execute-api:Invoke"]
    resources = ["execute-api:/${local.stage_name}/GET/health-check"]
  }
}

data "aws_iam_policy_document" "lpa_store_wildcard" {
  policy_id = "lpa-store-${var.environment_name}-${data.aws_region.current.name}-wildcard-resource-policy"

  statement {
    sid    = "AllowExecutionFromWildcards"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    actions   = ["execute-api:Invoke"]
    resources = ["*"]
    condition {
      test     = "ArnEquals"
      variable = "aws:PrincipalArn"
      values   = var.allowed_wildcard_arns
    }
  }
}

resource "aws_lambda_permission" "api_gateway_invoke" {
  for_each      = module.lambda
  statement_id  = "AllowLambdaAPIGatewayInvocation"
  action        = "lambda:InvokeFunction"
  function_name = each.value.function_name
  principal     = "apigateway.amazonaws.com"
  # The /* part allows invocation from any stage, method and resource path
  # within API Gateway.
  source_arn = "${aws_api_gateway_rest_api.lpa_store.execution_arn}/*"

  provider = aws.region
}

resource "aws_api_gateway_base_path_mapping" "mapping" {
  api_id      = aws_api_gateway_rest_api.lpa_store.id
  stage_name  = aws_api_gateway_stage.current.stage_name
  domain_name = aws_api_gateway_domain_name.lpa_store.domain_name

  lifecycle {
    create_before_destroy = true
  }

  provider = aws.region
}
