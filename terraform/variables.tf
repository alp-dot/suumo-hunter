variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ap-northeast-1"
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

variable "line_notify_token" {
  description = "LINE Notify API token"
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
  default     = "rate(1 hour)"
}
