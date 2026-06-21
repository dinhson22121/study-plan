###############################################################################
# EC2 app host (app + worker + nginx via the existing docker-compose prod
# override) in a PRIVATE subnet behind the ALB. user_data installs Docker +
# Compose, clones the repo to /opt/edu-app, and renders /opt/edu-app/.env from
# SSM before running deploy/scripts/deploy.sh.
#
# Note: nginx still terminates TLS inside compose per the runbook, but with the
# ALB primary approach TLS is also terminated at the ALB (ACM). The ALB forwards
# to the app over HTTP on 8080. If you instead want nginx to own TLS with an EIP,
# see deploy/terraform/README.md "Alternative: nginx on EC2 with an EIP".
###############################################################################

# Latest Amazon Linux 2023 AMI (x86_64).
data "aws_ami" "al2023" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-*-x86_64"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

locals {
  user_data = templatefile("${path.module}/templates/user_data.sh.tftpl", {
    aws_region      = var.aws_region
    ssm_path_prefix = var.ssm_path_prefix
    app_repo_url    = var.app_repo_url
    app_repo_branch = var.app_repo_branch
    env_name        = var.environment
  })
}

resource "aws_instance" "app" {
  ami                    = data.aws_ami.al2023.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.private[0].id
  vpc_security_group_ids = [aws_security_group.app.id]
  iam_instance_profile   = aws_iam_instance_profile.ec2.name
  key_name               = var.key_pair_name == "" ? null : var.key_pair_name

  user_data                   = local.user_data
  user_data_replace_on_change = true

  root_block_device {
    volume_type           = "gp3"
    volume_size           = var.root_volume_size_gb
    encrypted             = true
    delete_on_termination = true
  }

  metadata_options {
    http_tokens   = "required" # IMDSv2 only
    http_endpoint = "enabled"
  }

  monitoring = true

  tags = { Name = "${local.name}-app" }

  # SSM params + log group must exist before the host boots and pulls them.
  depends_on = [
    aws_ssm_parameter.edu_postgres_url,
    aws_ssm_parameter.edu_redis_url,
    aws_ssm_parameter.edu_jwt_secret,
    aws_cloudwatch_log_group.app,
  ]
}
