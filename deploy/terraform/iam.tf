###############################################################################
# IAM: EC2 instance role + profile with least-privilege access to:
#   - the S3 assets bucket (Get/Put/Delete/List on that bucket only),
#   - SSM Parameter Store read (decrypt) under the app's path prefix,
#   - CloudWatch Logs write,
#   - SSM Session Manager (AmazonSSMManagedInstanceCore) for keyless shell access,
#   - MSK cluster access (only when enable_msk).
###############################################################################

data "aws_iam_policy_document" "ec2_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "ec2" {
  name               = "${local.name}-ec2-role"
  assume_role_policy = data.aws_iam_policy_document.ec2_assume.json

  tags = { Name = "${local.name}-ec2-role" }
}

resource "aws_iam_instance_profile" "ec2" {
  name = "${local.name}-ec2-profile"
  role = aws_iam_role.ec2.name
}

# ---------------------------------------------------------------------------
# S3 access scoped to the assets bucket.
# ---------------------------------------------------------------------------

data "aws_iam_policy_document" "ec2_s3" {
  statement {
    sid       = "ListBucket"
    actions   = ["s3:ListBucket", "s3:GetBucketLocation"]
    resources = [aws_s3_bucket.assets.arn]
  }

  statement {
    sid       = "ObjectRW"
    actions   = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"]
    resources = ["${aws_s3_bucket.assets.arn}/*"]
  }
}

resource "aws_iam_role_policy" "ec2_s3" {
  name   = "${local.name}-s3"
  role   = aws_iam_role.ec2.id
  policy = data.aws_iam_policy_document.ec2_s3.json
}

# ---------------------------------------------------------------------------
# SSM Parameter Store read for the app's secret path (decrypt SecureStrings).
# ---------------------------------------------------------------------------

data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "ec2_ssm_params" {
  statement {
    sid     = "ReadAppParams"
    actions = ["ssm:GetParameter", "ssm:GetParameters", "ssm:GetParametersByPath"]
    resources = [
      "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter${var.ssm_path_prefix}",
      "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter${var.ssm_path_prefix}/*",
    ]
  }

  # SecureStrings are encrypted with the AWS-managed SSM key; allow decrypt for
  # SSM-issued grants only (least privilege via the kms:ViaService condition).
  statement {
    sid       = "DecryptSsmParams"
    actions   = ["kms:Decrypt"]
    resources = ["*"]

    condition {
      test     = "StringEquals"
      variable = "kms:ViaService"
      values   = ["ssm.${var.aws_region}.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy" "ec2_ssm_params" {
  name   = "${local.name}-ssm-params"
  role   = aws_iam_role.ec2.id
  policy = data.aws_iam_policy_document.ec2_ssm_params.json
}

# ---------------------------------------------------------------------------
# CloudWatch Logs write.
# ---------------------------------------------------------------------------

data "aws_iam_policy_document" "ec2_logs" {
  statement {
    sid = "CloudWatchLogs"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DescribeLogStreams",
    ]
    resources = ["${aws_cloudwatch_log_group.app.arn}:*"]
  }
}

resource "aws_iam_role_policy" "ec2_logs" {
  name   = "${local.name}-logs"
  role   = aws_iam_role.ec2.id
  policy = data.aws_iam_policy_document.ec2_logs.json
}

# ---------------------------------------------------------------------------
# SSM Session Manager (keyless shell). Managed policy is the AWS-recommended
# minimum for SSM-managed instances.
# ---------------------------------------------------------------------------

resource "aws_iam_role_policy_attachment" "ec2_ssm_core" {
  role       = aws_iam_role.ec2.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

# ---------------------------------------------------------------------------
# MSK cluster access (only when enable_msk). Scoped to this cluster ARN.
# ---------------------------------------------------------------------------

data "aws_iam_policy_document" "ec2_msk" {
  count = var.enable_msk ? 1 : 0

  statement {
    sid = "KafkaConnect"
    actions = [
      "kafka-cluster:Connect",
      "kafka-cluster:DescribeCluster",
      "kafka-cluster:AlterCluster",
    ]
    resources = [aws_msk_serverless_cluster.main[0].arn]
  }

  statement {
    sid = "KafkaTopicAndGroup"
    actions = [
      "kafka-cluster:*Topic*",
      "kafka-cluster:WriteData",
      "kafka-cluster:ReadData",
      "kafka-cluster:AlterGroup",
      "kafka-cluster:DescribeGroup",
    ]
    # Topic/group ARNs derive from the cluster ARN; wildcard the suffix.
    resources = [
      replace(aws_msk_serverless_cluster.main[0].arn, ":cluster/", ":topic/"),
      replace(aws_msk_serverless_cluster.main[0].arn, ":cluster/", ":group/"),
      "${replace(aws_msk_serverless_cluster.main[0].arn, ":cluster/", ":topic/")}/*",
      "${replace(aws_msk_serverless_cluster.main[0].arn, ":cluster/", ":group/")}/*",
    ]
  }
}

resource "aws_iam_role_policy" "ec2_msk" {
  count = var.enable_msk ? 1 : 0

  name   = "${local.name}-msk"
  role   = aws_iam_role.ec2.id
  policy = data.aws_iam_policy_document.ec2_msk[0].json
}
