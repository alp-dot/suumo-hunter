resource "aws_s3_bucket" "properties" {
  bucket = "${var.project_name}-properties-${data.aws_caller_identity.current.account_id}"
}

resource "aws_s3_bucket_versioning" "properties" {
  bucket = aws_s3_bucket.properties.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "properties" {
  bucket = aws_s3_bucket.properties.id

  rule {
    id     = "delete-old-versions"
    status = "Enabled"

    filter {}

    noncurrent_version_expiration {
      noncurrent_days = 30
    }
  }
}

resource "aws_s3_bucket_public_access_block" "properties" {
  bucket = aws_s3_bucket.properties.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

data "aws_caller_identity" "current" {}
