# EventBridge rule for scheduled execution
resource "aws_cloudwatch_event_rule" "schedule" {
  name                = "${local.name_prefix}-schedule"
  description         = "Trigger SUUMO Hunter Lambda (${var.instance_name}) on schedule"
  schedule_expression = var.schedule_expression

  tags = local.common_tags
}

# EventBridge target
resource "aws_cloudwatch_event_target" "lambda" {
  rule      = aws_cloudwatch_event_rule.schedule.name
  target_id = "${local.name_prefix}-lambda"
  arn       = aws_lambda_function.suumo_hunter.arn
}
