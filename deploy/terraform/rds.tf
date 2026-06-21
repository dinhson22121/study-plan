###############################################################################
# RDS PostgreSQL 16. Private subnets, not publicly accessible, automated backups,
# Multi-AZ toggle. The app's connection string MUST use sslmode=require — see the
# edu_postgres_url SSM parameter in ssm.tf which appends it.
###############################################################################

resource "aws_db_subnet_group" "main" {
  name       = "${local.name}-db-subnets"
  subnet_ids = aws_subnet.private[*].id

  tags = { Name = "${local.name}-db-subnets" }
}

# Enforce TLS at the server: rds.force_ssl = 1 rejects non-SSL connections, which
# pairs with the app refusing to boot without sslmode=require.
resource "aws_db_parameter_group" "main" {
  name        = "${local.name}-pg16"
  family      = "postgres16"
  description = "Edu App PostgreSQL 16 parameters (force SSL)."

  parameter {
    name  = "rds.force_ssl"
    value = "1"
  }

  tags = { Name = "${local.name}-pg16" }
}

resource "aws_db_instance" "main" {
  identifier     = "${local.name}-postgres"
  engine         = "postgres"
  engine_version = var.db_engine_version
  instance_class = var.db_instance_class

  allocated_storage     = var.db_allocated_storage
  max_allocated_storage = var.db_max_allocated_storage
  storage_type          = "gp3"
  storage_encrypted     = true

  db_name  = var.db_name
  username = var.db_username
  password = var.db_password
  port     = 5432

  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  parameter_group_name   = aws_db_parameter_group.main.name
  publicly_accessible    = false
  multi_az               = var.db_multi_az

  # Backups + snapshots.
  backup_retention_period   = var.db_backup_retention_days
  backup_window             = "17:00-18:00" # UTC (~midnight ICT)
  maintenance_window        = "sun:18:30-sun:19:30"
  copy_tags_to_snapshot     = true
  delete_automated_backups  = false
  skip_final_snapshot       = false
  final_snapshot_identifier = "${local.name}-postgres-final"
  deletion_protection       = true

  performance_insights_enabled = true
  auto_minor_version_upgrade   = true
  apply_immediately            = false

  tags = { Name = "${local.name}-postgres" }
}
