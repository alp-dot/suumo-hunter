# Lambda execution role (shared across all instances)
resource "aws_iam_role" "lambda" {
  count = var.create_iam_role ? 1 : 0
  name  = local.iam_role_name

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Project   = var.project_name
    ManagedBy = "terraform"
  }
}

# Data source for existing IAM role (when create_iam_role = false)
data "aws_iam_role" "lambda" {
  count = var.create_iam_role ? 0 : 1
  name  = local.iam_role_name
}

# CloudWatch Logs policy
resource "aws_iam_role_policy" "lambda_logs" {
  count = var.create_iam_role ? 1 : 0
  name  = "${var.project_name}-lambda-logs"
  role  = aws_iam_role.lambda[0].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:*:*:*"
      }
    ]
  })
}

# S3 access policy (wildcard to support all instances)
resource "aws_iam_role_policy" "lambda_s3" {
  count = var.create_iam_role ? 1 : 0
  name  = "${var.project_name}-lambda-s3"
  role  = aws_iam_role.lambda[0].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject"
        ]
        Resource = "arn:aws:s3:::${var.project_name}-*-properties-*/*"
      }
    ]
  })
}
