###############################################################################
# Edu App — Phase 1 production infrastructure (single region, cost-conscious).
#
# This module turns the prose runbook in deploy/README.md into terraform-applyable
# infrastructure: VPC, ALB, EC2 (Docker Compose host), RDS PostgreSQL, ElastiCache
# Redis, optional MSK, private S3 assets bucket, IAM, SSM secrets, Route53 + ACM,
# and CloudWatch logging/alarms.
#
# NO real secrets or account IDs live in this repo. Everything sensitive flows in
# through variables (sensitive = true) and lands in SSM Parameter Store as
# SecureString. The EC2 instance role reads those parameters at deploy time.
###############################################################################

terraform {
  # Pin a known-good range. 1.5 introduced the `import` block and stable defaults
  # this module relies on; cap below 2.0 to avoid surprise breaking changes.
  required_version = ">= 1.5.0, < 2.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.40"
    }
  }

  # ---------------------------------------------------------------------------
  # Remote state backend (RECOMMENDED for any shared/production use).
  #
  # Left commented so `terraform init` works out-of-the-box with local state for
  # a first read-through. Before applying for real, create the S3 bucket + a
  # DynamoDB lock table (see deploy/terraform/README.md) and uncomment this block.
  #
  # backend "s3" {
  #   bucket         = "edu-app-tfstate-prod"      # pre-created, versioned, SSE-on
  #   key            = "edu-app/prod/terraform.tfstate"
  #   region         = "ap-southeast-1"
  #   dynamodb_table = "edu-app-tflock"            # pre-created, PK = LockID (S)
  #   encrypt        = true
  # }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = local.common_tags
  }
}

# ACM certificates for an ALB must live in the same region as the ALB, so no
# us-east-1 aliased provider is needed here (that is only required for CloudFront).

data "aws_availability_zones" "available" {
  state = "available"
}

locals {
  name = "${var.project}-${var.environment}"

  common_tags = merge(
    {
      Project     = var.project
      Environment = var.environment
      ManagedBy   = "terraform"
    },
    var.extra_tags,
  )

  # Use the first two available AZs in the region for the 2-AZ Phase 1 layout.
  azs = slice(data.aws_availability_zones.available.names, 0, 2)
}
