###############################################################################
# ElastiCache Redis (single node OK for Phase 1), private, in-transit TLS enabled
# so the app can use a rediss:// URL (deploy/README.md).
###############################################################################

resource "aws_elasticache_subnet_group" "main" {
  name       = "${local.name}-redis-subnets"
  subnet_ids = aws_subnet.private[*].id

  tags = { Name = "${local.name}-redis-subnets" }
}

# Single-node replication group with TLS in transit. Using a replication group
# (rather than a bare cluster) gives a stable primary endpoint and an easy path
# to add replicas later without re-creating the resource.
resource "aws_elasticache_replication_group" "main" {
  replication_group_id = "${local.name}-redis"
  description          = "Edu App Redis (Phase 1 single node)."

  engine         = "redis"
  engine_version = var.redis_engine_version
  node_type      = var.redis_node_type
  port           = 6379

  num_cache_clusters = 1

  subnet_group_name  = aws_elasticache_subnet_group.main.name
  security_group_ids = [aws_security_group.redis.id]

  # In-transit TLS => app must use rediss://. No auth token in Phase 1; the SG
  # already restricts access to the app SG only. Add auth_token later if needed.
  transit_encryption_enabled = true
  at_rest_encryption_enabled = true

  automatic_failover_enabled = false
  multi_az_enabled           = false

  apply_immediately = false

  tags = { Name = "${local.name}-redis" }
}
