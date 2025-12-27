# EventBridge rule for scheduled execution
resource "aws_cloudwatch_event_rule" "schedule" {
  name                = "${var.project_name}-schedule"
  description         = "Trigger SUUMO Hunter Lambda on schedule"
  schedule_expression = var.schedule_expression
}

# EventBridge target
resource "aws_cloudwatch_event_target" "lambda" {
  rule      = aws_cloudwatch_event_rule.schedule.name
  target_id = "suumo-hunter-lambda"
  arn       = aws_lambda_function.suumo_hunter.arn
}
