###############################################################################
# Kafka — OPTIONAL via enable_msk.
#
# Phase 1 default (enable_msk = false): skip MSK entirely and self-host Kafka in
# the docker-compose KRaft container on the EC2 host. The app then talks to
# kafka:29092 inside the compose network (deploy/README.md self-host note).
#
# enable_msk = true: provision MSK Serverless with IAM auth on port 9098. The
# EC2 instance role is granted kafka-cluster:* on this cluster (see iam.tf), and
# EDU_KAFKA_BROKERS must be set to the bootstrap brokers (read at deploy time
# with `aws kafka get-bootstrap-brokers`; MSK Serverless brokers are not known
# until after creation, so they are surfaced as an output, not an SSM param).
###############################################################################

resource "aws_msk_serverless_cluster" "main" {
  count = var.enable_msk ? 1 : 0

  cluster_name = "${local.name}-kafka"

  vpc_config {
    subnet_ids         = aws_subnet.private[*].id
    security_group_ids = [aws_security_group.msk[0].id]
  }

  client_authentication {
    sasl {
      iam {
        enabled = true
      }
    }
  }

  tags = { Name = "${local.name}-kafka" }
}
