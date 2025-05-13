locals {
  stage_name = "current"
  template_file = templatefile("../../docs/openapi/openapi-aws.compiled.yaml", {
    lambda_create_invoke_arn  = module.lambda["create"].invoke_arn
    lambda_get_invoke_arn     = module.lambda["get"].invoke_arn
    lambda_update_invoke_arn  = module.lambda["update"].invoke_arn
    lambda_getlist_invoke_arn = module.lambda["getlist"].invoke_arn
  })
}

resource "aws_api_gateway_rest_api" "lpa_store" {
  name        = "lpa-store-${var.environment_name}"
  description = "API Gateway for LPA Store - ${var.environment_name}"
  body        = local.template_file
  policy      = data.aws_iam_policy_document.lpa_store.json

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  provider = aws.region
}


resource "aws_api_gateway_deployment" "lpa_store" {
  rest_api_id = aws_api_gateway_rest_api.lpa_store.id

  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_rest_api.lpa_store.body,
      data.aws_iam_policy_document.lpa_store.json,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }

  depends_on = [aws_api_gateway_rest_api.lpa_store]

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
      "\"httpMethod\":\"$context.httpMethod\",",
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
  policy_id = "lpa-store-${var.environment_name}-${data.aws_region.current.name}-resource-policy"
  override_policy_documents = concat(
    length(var.environment.allowed_wildcard_arns) > 0 ? [data.aws_iam_policy_document.lpa_store_wildcard.json] : [],
    local.ip_restrictions_enabled ? [data.aws_iam_policy_document.lpa_rest_api_ip_restriction_policy[0].json] : []
  )
  statement {
    sid    = "AllowExecutionFromAllowedARNs"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = var.environment.allowed_arns
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
    resources = ["arn:aws:execute-api:${data.aws_region.current.name}:${var.environment.account_id}:*/current/GET/health-check"]
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
      values   = var.environment.allowed_wildcard_arns
    }
  }
}

data "aws_iam_policy_document" "lpa_rest_api_ip_restriction_policy" {
  count = local.ip_restrictions_enabled ? 1 : 0
  statement {
    sid    = "DenyExecuteByNoneAllowedIPRanges"
    effect = "Deny"
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
    actions       = ["execute-api:Invoke"]
    not_resources = ["arn:aws:execute-api:eu-west-?:${var.environment.account_id}:*/*/*/health-check"]
    condition {
      test     = "NotIpAddress"
      variable = "aws:SourceIp"
      values   = sensitive(local.allow_list_mapping[var.environment.account_name])
    }
  }
}

module "allow_list" {
  source = "git@github.com:ministryofjustice/opg-terraform-aws-moj-ip-allow-list.git?ref=v3.4.0"
}

locals {
  allow_list_mapping = {
    development = concat(
      module.allow_list.use_an_lpa_development,
      module.allow_list.mr_lpa_development,
      module.allow_list.sirius_dev_allow_list,
    )
    preproduction = concat(
      module.allow_list.use_an_lpa_preproduction,
      module.allow_list.mr_lpa_preproduction,
      module.allow_list.sirius_pre_allow_list,
    )
    production = concat(
      module.allow_list.use_an_lpa_production,
      module.allow_list.mr_lpa_production,
      module.allow_list.sirius_prod_allow_list,
    )
  }
  ip_restrictions_enabled = contains(["preproduction", "production"], var.environment.account_name)
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
