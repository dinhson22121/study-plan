###############################################################################
# Outputs. Sensitive values are marked so they don't print in plan/apply logs.
###############################################################################

output "vpc_id" {
  description = "VPC ID."
  value       = aws_vpc.main.id
}

output "public_subnet_ids" {
  description = "Public subnet IDs."
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "Private subnet IDs."
  value       = aws_subnet.private[*].id
}

output "alb_dns_name" {
  description = "ALB DNS name (target of the Route53 alias)."
  value       = aws_lb.main.dns_name
}

output "api_url" {
  description = "Public HTTPS URL for the API."
  value       = "https://${local.api_fqdn}"
}

output "route53_zone_id" {
  description = "Hosted zone ID used for DNS records."
  value       = local.zone_id
}

output "ec2_instance_id" {
  description = "App EC2 instance ID (use with SSM Session Manager)."
  value       = aws_instance.app.id
}

output "ec2_private_ip" {
  description = "App EC2 private IP."
  value       = aws_instance.app.private_ip
}

output "rds_endpoint" {
  description = "RDS endpoint host."
  value       = aws_db_instance.main.address
}

output "rds_port" {
  description = "RDS port."
  value       = aws_db_instance.main.port
}

output "redis_primary_endpoint" {
  description = "ElastiCache Redis primary endpoint."
  value       = aws_elasticache_replication_group.main.primary_endpoint_address
}

output "s3_bucket_name" {
  description = "Assets S3 bucket name."
  value       = aws_s3_bucket.assets.bucket
}

output "ec2_instance_role_arn" {
  description = "EC2 instance role ARN (referenced by the S3 bucket policy)."
  value       = aws_iam_role.ec2.arn
}

output "ssm_path_prefix" {
  description = "SSM Parameter Store path the host reads its .env from."
  value       = var.ssm_path_prefix
}

output "cloudwatch_log_group" {
  description = "App CloudWatch log group name."
  value       = aws_cloudwatch_log_group.app.name
}

output "msk_bootstrap_brokers_hint" {
  description = "How to fetch MSK bootstrap brokers when enable_msk = true."
  value = var.enable_msk ? (
    "Run: aws kafka get-bootstrap-brokers --cluster-arn ${aws_msk_serverless_cluster.main[0].arn} --query BootstrapBrokerStringSaslIam --output text  then put it in ${var.ssm_path_prefix}/EDU_KAFKA_BROKERS"
  ) : "MSK disabled (self-hosting Kafka on the EC2 compose stack)."
}

# Sensitive derived URLs — handy for the restore test / debugging, hidden from logs.
output "postgres_url" {
  description = "Full PostgreSQL connection URL (sslmode=require)."
  value       = local.postgres_url
  sensitive   = true
}

output "redis_url" {
  description = "Full Redis connection URL (rediss:// TLS)."
  value       = local.redis_url
  sensitive   = true
}
