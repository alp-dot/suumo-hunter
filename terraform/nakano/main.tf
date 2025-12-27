terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # 本番運用時はS3バックエンドを設定
  # backend "s3" {
  #   bucket = "your-terraform-state-bucket"
  #   key    = "suumo-hunter/shibuya/terraform.tfstate"
  #   region = "ap-northeast-1"
  # }
}

provider "aws" {
  region = "ap-northeast-1"
}

module "suumo_hunter" {
  source = "../modules/suumo-hunter"

  instance_name       = var.instance_name
  suumo_search_url    = var.suumo_search_url
  discord_webhook_url = var.discord_webhook_url
  max_page            = var.max_page
  schedule_expression = var.schedule_expression
  create_iam_role     = var.create_iam_role
  lambda_zip_path     = "${path.module}/../../build/lambda.zip"
}

# Variables
variable "instance_name" {
  type = string
}

variable "suumo_search_url" {
  type = string
}

variable "discord_webhook_url" {
  type      = string
  sensitive = true
}

variable "max_page" {
  type    = number
  default = 30
}

variable "schedule_expression" {
  type    = string
  default = "cron(15 0,6,9,13 * * ? *)"
}

variable "create_iam_role" {
  type    = bool
  default = true
}

# Outputs
output "lambda_function_name" {
  value = module.suumo_hunter.lambda_function_name
}

output "s3_bucket_name" {
  value = module.suumo_hunter.s3_bucket_name
}
