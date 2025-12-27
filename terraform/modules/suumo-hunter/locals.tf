locals {
  # Resource name prefix: suumo-hunter-{instance_name}
  name_prefix = "${var.project_name}-${var.instance_name}"

  # Common IAM role name (shared across instances)
  iam_role_name = "${var.project_name}-lambda-role"

  # Lambda role ARN (use created or existing)
  lambda_role_arn = var.create_iam_role ? aws_iam_role.lambda[0].arn : data.aws_iam_role.lambda[0].arn

  # Common tags for all resources
  common_tags = {
    Project   = var.project_name
    Instance  = var.instance_name
    ManagedBy = "terraform"
  }
}
