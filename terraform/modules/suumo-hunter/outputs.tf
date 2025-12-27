output "lambda_function_name" {
  description = "Name of the Lambda function"
  value       = aws_lambda_function.suumo_hunter.function_name
}

output "lambda_function_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.suumo_hunter.arn
}

output "s3_bucket_name" {
  description = "Name of the S3 bucket"
  value       = aws_s3_bucket.properties.id
}

output "eventbridge_rule_name" {
  description = "Name of the EventBridge rule"
  value       = aws_cloudwatch_event_rule.schedule.name
}

output "iam_role_arn" {
  description = "ARN of the IAM role (if created)"
  value       = var.create_iam_role ? aws_iam_role.lambda[0].arn : data.aws_iam_role.lambda[0].arn
}
