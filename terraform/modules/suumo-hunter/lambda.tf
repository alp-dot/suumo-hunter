# Lambda function
resource "aws_lambda_function" "suumo_hunter" {
  function_name = local.name_prefix
  role          = local.lambda_role_arn

  filename         = var.lambda_zip_path
  source_code_hash = filebase64sha256(var.lambda_zip_path)

  handler       = "bootstrap"
  runtime       = "provided.al2023"
  architectures = ["arm64"]
  timeout       = 300
  memory_size   = 256

  environment {
    variables = {
      BUCKET_NAME         = aws_s3_bucket.properties.id
      BUCKET_KEY          = "properties.csv"
      MAX_PAGE            = tostring(var.max_page)
      SUUMO_SEARCH_URL    = var.suumo_search_url
      DISCORD_WEBHOOK_URL = var.discord_webhook_url
    }
  }

  tags = local.common_tags
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${local.name_prefix}"
  retention_in_days = 14

  tags = local.common_tags
}

# Lambda permission for EventBridge
resource "aws_lambda_permission" "eventbridge" {
  statement_id  = "AllowEventBridgeInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.suumo_hunter.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.schedule.arn
}
