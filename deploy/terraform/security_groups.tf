###############################################################################
# Security groups (least privilege).
#   alb   : 80/443 from the public internet.
#   app   : 8080 from the ALB only (+ optional SSH from admin CIDR).
#   rds   : 5432 from the app SG only.
#   redis : 6379 from the app SG only.
#   msk   : 9098 (IAM auth) from the app SG only (created only when enable_msk).
# No data-store SG opens to 0.0.0.0/0 (deploy/README.md).
###############################################################################

# ---------------------------------------------------------------------------
# ALB
# ---------------------------------------------------------------------------

resource "aws_security_group" "alb" {
  name        = "${local.name}-alb-sg"
  description = "Public ALB ingress (80/443)."
  vpc_id      = aws_vpc.main.id

  tags = { Name = "${local.name}-alb-sg" }
}

resource "aws_vpc_security_group_ingress_rule" "alb_http" {
  security_group_id = aws_security_group.alb.id
  description       = "HTTP (redirected to HTTPS)."
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 80
  to_port           = 80
  ip_protocol       = "tcp"
}

resource "aws_vpc_security_group_ingress_rule" "alb_https" {
  security_group_id = aws_security_group.alb.id
  description       = "HTTPS."
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 443
  to_port           = 443
  ip_protocol       = "tcp"
}

resource "aws_vpc_security_group_egress_rule" "alb_all" {
  security_group_id = aws_security_group.alb.id
  description       = "Allow all egress (to app targets)."
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
}

# ---------------------------------------------------------------------------
# App / EC2
# ---------------------------------------------------------------------------

resource "aws_security_group" "app" {
  name        = "${local.name}-app-sg"
  description = "App EC2 host: traffic from ALB only, all egress."
  vpc_id      = aws_vpc.main.id

  tags = { Name = "${local.name}-app-sg" }
}

resource "aws_vpc_security_group_ingress_rule" "app_from_alb" {
  security_group_id            = aws_security_group.app.id
  description                  = "App HTTP port from the ALB only."
  referenced_security_group_id = aws_security_group.alb.id
  from_port                    = 8080
  to_port                      = 8080
  ip_protocol                  = "tcp"
}

# Optional SSH from a single admin CIDR. Prefer SSM Session Manager and leave
# admin_ssh_cidr empty in production.
resource "aws_vpc_security_group_ingress_rule" "app_ssh" {
  count = var.admin_ssh_cidr == "" ? 0 : 1

  security_group_id = aws_security_group.app.id
  description       = "SSH from admin CIDR."
  cidr_ipv4         = var.admin_ssh_cidr
  from_port         = 22
  to_port           = 22
  ip_protocol       = "tcp"
}

resource "aws_vpc_security_group_egress_rule" "app_all" {
  security_group_id = aws_security_group.app.id
  description       = "Allow all egress (NAT, RDS, Redis, S3, SSM, ECR, MSK)."
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
}

# ---------------------------------------------------------------------------
# RDS
# ---------------------------------------------------------------------------

resource "aws_security_group" "rds" {
  name        = "${local.name}-rds-sg"
  description = "RDS PostgreSQL: 5432 from app SG only."
  vpc_id      = aws_vpc.main.id

  tags = { Name = "${local.name}-rds-sg" }
}

resource "aws_vpc_security_group_ingress_rule" "rds_from_app" {
  security_group_id            = aws_security_group.rds.id
  description                  = "PostgreSQL from app SG."
  referenced_security_group_id = aws_security_group.app.id
  from_port                    = 5432
  to_port                      = 5432
  ip_protocol                  = "tcp"
}

# ---------------------------------------------------------------------------
# ElastiCache Redis
# ---------------------------------------------------------------------------

resource "aws_security_group" "redis" {
  name        = "${local.name}-redis-sg"
  description = "ElastiCache Redis: 6379 from app SG only."
  vpc_id      = aws_vpc.main.id

  tags = { Name = "${local.name}-redis-sg" }
}

resource "aws_vpc_security_group_ingress_rule" "redis_from_app" {
  security_group_id            = aws_security_group.redis.id
  description                  = "Redis from app SG."
  referenced_security_group_id = aws_security_group.app.id
  from_port                    = 6379
  to_port                      = 6379
  ip_protocol                  = "tcp"
}

# ---------------------------------------------------------------------------
# MSK (only when enable_msk)
# ---------------------------------------------------------------------------

resource "aws_security_group" "msk" {
  count = var.enable_msk ? 1 : 0

  name        = "${local.name}-msk-sg"
  description = "MSK Serverless: 9098 (IAM auth) from app SG only."
  vpc_id      = aws_vpc.main.id

  tags = { Name = "${local.name}-msk-sg" }
}

resource "aws_vpc_security_group_ingress_rule" "msk_from_app" {
  count = var.enable_msk ? 1 : 0

  security_group_id            = aws_security_group.msk[0].id
  description                  = "Kafka IAM-auth port from app SG."
  referenced_security_group_id = aws_security_group.app.id
  from_port                    = 9098
  to_port                      = 9098
  ip_protocol                  = "tcp"
}

resource "aws_vpc_security_group_egress_rule" "msk_all" {
  count = var.enable_msk ? 1 : 0

  security_group_id = aws_security_group.msk[0].id
  description       = "Allow all egress."
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
}
