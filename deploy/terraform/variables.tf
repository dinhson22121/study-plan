###############################################################################
# Input variables. Secrets are marked `sensitive = true` and must be supplied at
# apply time (env var TF_VAR_*, a gitignored *.auto.tfvars, or -var) — NEVER
# hardcoded. See terraform.tfvars.example for the full placeholder set.
###############################################################################

# ---------------------------------------------------------------------------
# Core / tagging
# ---------------------------------------------------------------------------

variable "project" {
  description = "Project slug used for naming and tags."
  type        = string
  default     = "edu-app"
}

variable "environment" {
  description = "Deployment environment (drives naming + the EDU_ENV the app runs as)."
  type        = string
  default     = "prod"
}

variable "aws_region" {
  description = "AWS region for all resources."
  type        = string
  default     = "ap-southeast-1"
}

variable "extra_tags" {
  description = "Additional tags merged into the common tag map on every resource."
  type        = map(string)
  default     = {}
}

# ---------------------------------------------------------------------------
# Networking
# ---------------------------------------------------------------------------

variable "vpc_cidr" {
  description = "CIDR block for the VPC."
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for the public subnets (one per AZ). Must match the 2-AZ layout."
  type        = list(string)
  default     = ["10.0.0.0/24", "10.0.1.0/24"]
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for the private subnets (one per AZ)."
  type        = list(string)
  default     = ["10.0.10.0/24", "10.0.11.0/24"]
}

variable "single_nat_gateway" {
  description = "Use a single NAT gateway (cheaper) instead of one per AZ (more resilient). Phase 1 default: single."
  type        = bool
  default     = true
}

variable "admin_ssh_cidr" {
  description = "CIDR allowed to SSH (22) to the EC2 host. Set to your admin IP/32. Empty disables SSH ingress."
  type        = string
  default     = ""
}

# ---------------------------------------------------------------------------
# Compute (EC2)
# ---------------------------------------------------------------------------

variable "instance_type" {
  description = "EC2 instance type for the Docker Compose app host. README suggests t3.large to start."
  type        = string
  default     = "t3.large"
}

variable "root_volume_size_gb" {
  description = "Root EBS (gp3) volume size in GB."
  type        = number
  default     = 50
}

variable "key_pair_name" {
  description = "Existing EC2 key pair name for SSH. Empty = no key pair attached (use SSM Session Manager)."
  type        = string
  default     = ""
}

variable "app_repo_url" {
  description = "Git URL the EC2 user_data clones to /opt/edu-app."
  type        = string
  default     = "https://github.com/your-org/edu-app.git"
}

variable "app_repo_branch" {
  description = "Git branch/tag to check out on the EC2 host."
  type        = string
  default     = "main"
}

# ---------------------------------------------------------------------------
# RDS PostgreSQL
# ---------------------------------------------------------------------------

variable "db_engine_version" {
  description = "PostgreSQL engine version. README requires PostgreSQL 16."
  type        = string
  default     = "16.4"
}

variable "db_instance_class" {
  description = "RDS instance class."
  type        = string
  default     = "db.t3.medium"
}

variable "db_allocated_storage" {
  description = "Initial RDS storage in GB."
  type        = number
  default     = 50
}

variable "db_max_allocated_storage" {
  description = "Upper bound for RDS storage autoscaling in GB."
  type        = number
  default     = 200
}

variable "db_multi_az" {
  description = "Enable RDS Multi-AZ. Toggle on for production resilience; off saves cost in early Phase 1."
  type        = bool
  default     = true
}

variable "db_backup_retention_days" {
  description = "Automated backup retention in days (README: >= 7)."
  type        = number
  default     = 7

  validation {
    condition     = var.db_backup_retention_days >= 7
    error_message = "Backup retention must be at least 7 days (deploy/README.md §4)."
  }
}

variable "db_name" {
  description = "Initial database name."
  type        = string
  default     = "eduapp"
}

variable "db_username" {
  description = "Master DB username."
  type        = string
  default     = "eduapp"
}

variable "db_password" {
  description = "Master DB password. Provide via TF_VAR_db_password — never commit."
  type        = string
  sensitive   = true
}

# ---------------------------------------------------------------------------
# ElastiCache Redis
# ---------------------------------------------------------------------------

variable "redis_engine_version" {
  description = "ElastiCache Redis engine version."
  type        = string
  default     = "7.1"
}

variable "redis_node_type" {
  description = "ElastiCache node type. Single node is acceptable for Phase 1."
  type        = string
  default     = "cache.t3.micro"
}

# ---------------------------------------------------------------------------
# MSK (Kafka) — optional
# ---------------------------------------------------------------------------

variable "enable_msk" {
  description = "Provision MSK Serverless for Kafka. Phase 1 may self-host Kafka on the EC2 (set false)."
  type        = bool
  default     = false
}

# ---------------------------------------------------------------------------
# S3 assets bucket
# ---------------------------------------------------------------------------

variable "s3_bucket_name" {
  description = "Globally-unique S3 bucket name for app assets."
  type        = string
  default     = "edu-assets-prod"
}

variable "s3_noncurrent_version_expiration_days" {
  description = "Expire noncurrent S3 object versions after N days (lifecycle rule)."
  type        = number
  default     = 30
}

# ---------------------------------------------------------------------------
# DNS + TLS
# ---------------------------------------------------------------------------

variable "domain_name" {
  description = "Apex/parent domain that owns the Route53 hosted zone, e.g. example.com."
  type        = string
}

variable "api_subdomain" {
  description = "Subdomain for the API host (joined with domain_name), e.g. 'api' => api.example.com."
  type        = string
  default     = "api"
}

variable "create_route53_zone" {
  description = "Create the hosted zone here (true) or look up an existing one by domain_name (false)."
  type        = bool
  default     = false
}

# ---------------------------------------------------------------------------
# Application secrets -> SSM Parameter Store (SecureString). All sensitive.
# ---------------------------------------------------------------------------

variable "ssm_path_prefix" {
  description = "SSM Parameter Store path prefix for app secrets, e.g. /edu-app/prod."
  type        = string
  default     = "/edu-app/prod"
}

variable "jwt_secret" {
  description = "EDU_JWT_SECRET — >= 32 chars, non-default. `openssl rand -base64 48`."
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.jwt_secret) >= 32
    error_message = "jwt_secret must be at least 32 characters (config.validateProduction enforces this)."
  }
}

variable "sentry_dsn" {
  description = "EDU_SENTRY_DSN — optional. Empty disables Sentry."
  type        = string
  default     = ""
  sensitive   = true
}

# ---------------------------------------------------------------------------
# CloudWatch alarms
# ---------------------------------------------------------------------------

variable "log_retention_days" {
  description = "CloudWatch Logs retention in days."
  type        = number
  default     = 30
}

variable "alarm_sns_topic_arn" {
  description = "Existing SNS topic ARN to notify on alarms. Empty = alarms created without an action."
  type        = string
  default     = ""
}

variable "alb_5xx_threshold" {
  description = "ALB 5xx count over the evaluation period that trips the alarm."
  type        = number
  default     = 10
}

variable "ec2_cpu_threshold_percent" {
  description = "EC2 average CPU percent that trips the high-CPU alarm."
  type        = number
  default     = 80
}
