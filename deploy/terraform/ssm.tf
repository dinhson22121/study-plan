###############################################################################
# SSM Parameter Store (SecureString) for app secrets/config.
#
# Secret flow: terraform writes these parameters from sensitive variables and
# from resource attributes (RDS endpoint, Redis endpoint). On the EC2 host the
# deploy renders /opt/edu-app/.env from this path with `aws ssm
# get-parameters-by-path --with-decryption` (deploy/README.md §3), which the
# docker-compose prod override consumes.
#
# Parameter names are UPPER_SNAKE so the README's awk snippet maps them straight
# to EDU_* env keys (it uppercases the leaf name).
###############################################################################

# Derived connection strings (built here so .env always has sslmode=require and
# rediss:// TLS, matching config.validateProduction + the runbook).
locals {
  postgres_url = format(
    "postgres://%s:%s@%s:%d/%s?sslmode=require",
    var.db_username,
    var.db_password,
    aws_db_instance.main.address,
    aws_db_instance.main.port,
    var.db_name,
  )

  # ElastiCache in-transit TLS => rediss://. primary_endpoint_address is the
  # writer endpoint of the single-node replication group.
  redis_url = format(
    "rediss://%s:%d/0",
    aws_elasticache_replication_group.main.primary_endpoint_address,
    6379,
  )
}

# ---- Secrets (SecureString) ----

resource "aws_ssm_parameter" "edu_jwt_secret" {
  name        = "${var.ssm_path_prefix}/EDU_JWT_SECRET"
  description = "JWT signing secret."
  type        = "SecureString"
  value       = var.jwt_secret

  tags = { Name = "${local.name}-jwt-secret" }
}

resource "aws_ssm_parameter" "db_password" {
  name        = "${var.ssm_path_prefix}/POSTGRES_PASSWORD"
  description = "RDS master password (also embedded in EDU_POSTGRES_URL)."
  type        = "SecureString"
  value       = var.db_password

  tags = { Name = "${local.name}-db-password" }
}

resource "aws_ssm_parameter" "edu_postgres_url" {
  name        = "${var.ssm_path_prefix}/EDU_POSTGRES_URL"
  description = "RDS PostgreSQL URL (sslmode=require)."
  type        = "SecureString"
  value       = local.postgres_url

  tags = { Name = "${local.name}-postgres-url" }
}

resource "aws_ssm_parameter" "edu_redis_url" {
  name        = "${var.ssm_path_prefix}/EDU_REDIS_URL"
  description = "ElastiCache Redis URL (rediss:// TLS)."
  type        = "SecureString"
  value       = local.redis_url

  tags = { Name = "${local.name}-redis-url" }
}

resource "aws_ssm_parameter" "edu_sentry_dsn" {
  name        = "${var.ssm_path_prefix}/EDU_SENTRY_DSN"
  description = "Sentry DSN (empty disables Sentry)."
  type        = "SecureString"
  value       = var.sentry_dsn == "" ? "disabled" : var.sentry_dsn

  tags = { Name = "${local.name}-sentry-dsn" }
}

# ---- Non-secret config (String). Kept in the same path so one fetch hydrates
#       the whole .env. S3 access is via the instance role, so no S3 keys here. ----

resource "aws_ssm_parameter" "edu_s3_bucket" {
  name  = "${var.ssm_path_prefix}/EDU_S3_BUCKET"
  type  = "String"
  value = aws_s3_bucket.assets.bucket

  tags = { Name = "${local.name}-s3-bucket" }
}

resource "aws_ssm_parameter" "edu_s3_region" {
  name  = "${var.ssm_path_prefix}/EDU_S3_REGION"
  type  = "String"
  value = var.aws_region

  tags = { Name = "${local.name}-s3-region" }
}

# Empty endpoint => the AWS SDK targets real S3 in EDU_S3_REGION. Stored as a
# single space because SSM rejects empty String values; the app trims it.
resource "aws_ssm_parameter" "edu_s3_endpoint" {
  name  = "${var.ssm_path_prefix}/EDU_S3_ENDPOINT"
  type  = "String"
  value = " "

  tags = { Name = "${local.name}-s3-endpoint" }
}

resource "aws_ssm_parameter" "edu_kafka_brokers" {
  name        = "${var.ssm_path_prefix}/EDU_KAFKA_BROKERS"
  description = "Kafka bootstrap brokers. Self-host default points at the on-box KRaft container."
  type        = "String"
  # MSK Serverless bootstrap brokers are not known at apply time; fetch them on
  # the host with `aws kafka get-bootstrap-brokers` and overwrite this value.
  # Self-host (enable_msk=false) uses the compose network broker.
  value = var.enable_msk ? "SET_FROM_get-bootstrap-brokers" : "kafka:29092"

  tags = { Name = "${local.name}-kafka-brokers" }

  lifecycle {
    # When MSK is on, the real broker list is set out-of-band after creation;
    # don't let terraform clobber it on subsequent applies.
    ignore_changes = [value]
  }
}
