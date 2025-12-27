variable "instance_name" {
  description = "Instance name for this deployment (e.g., shibuya, shinjuku)"
  type        = string
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "suumo-hunter"
}

variable "suumo_search_url" {
  description = "SUUMO search URL to scrape"
  type        = string
}

variable "discord_webhook_url" {
  description = "Discord Webhook URL for notifications"
  type        = string
  sensitive   = true
}

variable "max_page" {
  description = "Maximum number of SUUMO pages to scrape"
  type        = number
  default     = 30
}

variable "schedule_expression" {
  description = "EventBridge schedule expression (cron or rate)"
  type        = string
  default     = "cron(15 0,6,9,13 * * ? *)" # JST 09:15, 15:15, 18:15, 22:15
}

variable "create_iam_role" {
  description = "Whether to create IAM role (set to false for additional instances to share existing role)"
  type        = bool
  default     = true
}

variable "lambda_zip_path" {
  description = "Path to the Lambda zip file"
  type        = string
}
