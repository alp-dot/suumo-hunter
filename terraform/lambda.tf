# Lambda function
resource "aws_lambda_function" "suumo_hunter" {
  function_name = var.project_name
  role          = aws_iam_role.lambda.arn

  filename         = data.archive_file.lambda_zip.output_path
  source_code_hash = data.archive_file.lambda_zip.output_base64sha256

  handler     = "bootstrap"
  runtime     = "provided.al2023"
  architectures = ["arm64"]
  timeout     = 300
  memory_size = 256

  environment {
    variables = {
      BUCKET_NAME       = aws_s3_bucket.properties.id
      BUCKET_KEY        = "properties.csv"
      MAX_PAGE          = tostring(var.max_page)
      SUUMO_SEARCH_URL  = var.suumo_search_url
      LINE_NOTIFY_TOKEN = var.line_notify_token
    }
  }
}

# Create zip from bootstrap binary
data "archive_file" "lambda_zip" {
  type        = "zip"
  source_file = "${path.module}/../build/bootstrap"
  output_path = "${path.module}/../build/lambda.zip"
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${var.project_name}"
  retention_in_days = 14
}

# Lambda permission for EventBridge
resource "aws_lambda_permission" "eventbridge" {
  statement_id  = "AllowEventBridgeInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.suumo_hunter.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.schedule.arn
}
