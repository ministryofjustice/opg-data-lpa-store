resource "aws_cloudwatch_event_rule" "metric_events" {
  name           = "${data.aws_default_tags.current.tags.environment-name}-metric-events"
  description    = "forward events to opg-metrics service"
  event_bus_name = aws_cloudwatch_event_bus.main.name

  event_pattern = jsonencode({
    source      = ["opg.poas.lpastore"]
    detail-type = ["metric"]
  })
  provider = aws.eu_west_1
}

data "aws_ssm_parameter" "opg_metrics_arn" {
  name     = "opg-metrics-api-destination-arn"
  provider = aws.eu_west_1
}

resource "aws_iam_role_policy" "opg_metrics" {
  name     = "opg-metrics-${data.aws_region.current.name}"
  role     = aws_iam_role.opg_metrics.name
  policy   = data.aws_iam_policy_document.opg_metrics.json
  provider = aws.eu_west_1
}


data "aws_iam_policy_document" "opg_metrics" {
  statement {
    effect  = "Allow"
    actions = ["events:InvokeApiDestination"]
    resources = [
      "${data.aws_ssm_parameter.opg_metrics_arn.value}*"
    ]
  }
  provider = aws.global
}

resource "aws_cloudwatch_event_target" "opg_metrics" {
  arn            = data.aws_ssm_parameter.opg_metrics_arn.insecure_value
  event_bus_name = aws_cloudwatch_event_bus.main.name
  rule           = aws_cloudwatch_event_rule.metric_events.name
  role_arn       = aws_iam_role.opg_metrics.arn
  http_target {
    header_parameters = {
      Content-Type = "application/json"
    }
    path_parameter_values   = []
    query_string_parameters = {}
  }
  input_path = "$.detail"
  provider   = aws.eu_west_1
}
