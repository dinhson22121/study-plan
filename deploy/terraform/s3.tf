###############################################################################
# S3 assets bucket: private (block all public access), versioning, SSE, lifecycle,
# and a least-privilege bucket policy (deny insecure transport + allow only the
# EC2 instance role). Mirrors the sample policy in deploy/README.md §1.
###############################################################################

resource "aws_s3_bucket" "assets" {
  bucket = var.s3_bucket_name

  tags = { Name = var.s3_bucket_name }
}

resource "aws_s3_bucket_public_access_block" "assets" {
  bucket = aws_s3_bucket.assets.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_ownership_controls" "assets" {
  bucket = aws_s3_bucket.assets.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_versioning" "assets" {
  bucket = aws_s3_bucket.assets.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "assets" {
  bucket = aws_s3_bucket.assets.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256" # SSE-S3. Swap to aws:kms + a CMK if compliance requires.
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "assets" {
  bucket = aws_s3_bucket.assets.id

  # Required when at least one rule has an empty filter (apply to whole bucket).
  depends_on = [aws_s3_bucket_versioning.assets]

  rule {
    id     = "expire-noncurrent-versions"
    status = "Enabled"

    filter {} # whole bucket

    noncurrent_version_expiration {
      noncurrent_days = var.s3_noncurrent_version_expiration_days
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}

# ---------------------------------------------------------------------------
# Bucket policy: deny non-TLS, allow only the EC2 instance role.
# ---------------------------------------------------------------------------

data "aws_iam_policy_document" "assets_bucket" {
  statement {
    sid    = "DenyInsecureTransport"
    effect = "Deny"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions   = ["s3:*"]
    resources = [aws_s3_bucket.assets.arn, "${aws_s3_bucket.assets.arn}/*"]

    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
  }

  statement {
    sid    = "AllowAppRole"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = [aws_iam_role.ec2.arn]
    }

    actions   = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject", "s3:ListBucket"]
    resources = [aws_s3_bucket.assets.arn, "${aws_s3_bucket.assets.arn}/*"]
  }
}

resource "aws_s3_bucket_policy" "assets" {
  bucket = aws_s3_bucket.assets.id
  policy = data.aws_iam_policy_document.assets_bucket.json

  # The policy references the public access block; apply that first so we never
  # briefly expose the bucket.
  depends_on = [aws_s3_bucket_public_access_block.assets]
}
