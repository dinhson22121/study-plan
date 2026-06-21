# Edu App — Terraform (Phase 1 production)

Terraform IaC that provisions the AWS infrastructure described in
[`../README.md`](../README.md) (Workstream C). It stands up a single-region,
cost-conscious but production-shaped stack: VPC, ALB, an EC2 Docker-Compose host,
RDS PostgreSQL 16, ElastiCache Redis, an optional MSK Serverless cluster, a
private S3 assets bucket, IAM, SSM Parameter Store secrets, Route53 + ACM, and
CloudWatch logging/alarms.

> This module **provisions infra and bootstraps the EC2 host**. The application
> itself is still deployed by the existing `deploy/scripts/deploy.sh` running on
> that host (Docker Compose with `deploy/docker-compose.prod.yml`). Terraform
> writes the secrets into SSM; the host reads them into `/opt/edu-app/.env`.

---

## Architecture summary

| Layer | Resource | Notes |
|-------|----------|-------|
| Network | VPC + 2 public + 2 private subnets, IGW, NAT | `single_nat_gateway` toggles 1 vs per-AZ NAT |
| Edge | ALB (HTTPS + HTTP→HTTPS redirect) | ACM cert (DNS-validated), TLS 1.2/1.3 policy |
| Compute | 1 EC2 (`t3.large`) in a **private** subnet | runs app+worker+nginx via compose; `user_data` installs Docker, clones repo, renders `.env` from SSM, runs `deploy.sh` |
| Data | RDS PostgreSQL 16 (Multi-AZ toggle) | private, `rds.force_ssl=1`, encrypted, automated backups, deletion protection |
| Cache | ElastiCache Redis 7 (single node) | private, in-transit + at-rest TLS → app uses `rediss://` |
| Kafka | MSK Serverless (**optional**) | `enable_msk=false` ⇒ self-host KRaft in compose |
| Assets | S3 bucket | block public access, versioning, SSE, lifecycle, least-priv bucket policy |
| Identity | EC2 instance role/profile | S3 (bucket-scoped), SSM param read+decrypt, CW Logs, SSM Session Manager, MSK (when enabled) |
| Secrets | SSM Parameter Store (SecureString) | `EDU_JWT_SECRET`, `EDU_POSTGRES_URL`, `EDU_REDIS_URL`, `POSTGRES_PASSWORD`, `EDU_SENTRY_DSN` (+ non-secret config) |
| DNS/TLS | Route53 record + ACM | `api.<domain>` ALIAS → ALB |
| Observability | CloudWatch log group + 3 alarms | ALB 5xx, EC2 CPU, zero healthy hosts |

---

## Prerequisites

1. **Terraform** `>= 1.5.0, < 2.0.0` and the **AWS provider `~> 5.40`**.
2. **AWS credentials** with permissions to create the above (admin or an
   equivalent scoped role). `aws sts get-caller-identity` should work.
3. A **registered domain** with its Route53 hosted zone either already in the
   account (`create_route53_zone = false`, default — the zone is looked up) or
   created by this module (`create_route53_zone = true`). If the domain is
   registered elsewhere, delegate its NS records to this zone first.
4. **Remote state backend (recommended):** an S3 bucket + DynamoDB lock table,
   created once, out-of-band:

   ```sh
   aws s3api create-bucket --bucket edu-app-tfstate-prod \
     --region ap-southeast-1 \
     --create-bucket-configuration LocationConstraint=ap-southeast-1
   aws s3api put-bucket-versioning --bucket edu-app-tfstate-prod \
     --versioning-configuration Status=Enabled
   aws s3api put-bucket-encryption --bucket edu-app-tfstate-prod \
     --server-side-encryption-configuration \
     '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}'
   aws dynamodb create-table --table-name edu-app-tflock \
     --attribute-definitions AttributeName=LockID,AttributeType=S \
     --key-schema AttributeName=LockID,KeyType=HASH \
     --billing-mode PAY_PER_REQUEST --region ap-southeast-1
   ```

   Then uncomment the `backend "s3"` block in `main.tf` and run
   `terraform init -migrate-state`.

---

## Usage

```sh
cd deploy/terraform

# 1. Provide secrets via environment (never commit them):
export TF_VAR_db_password="$(openssl rand -base64 24)"
export TF_VAR_jwt_secret="$(openssl rand -base64 48)"   # >= 32 chars (validated)
export TF_VAR_sentry_dsn=""                              # optional

# 2. Copy + edit the non-secret variables:
cp terraform.tfvars.example terraform.tfvars
$EDITOR terraform.tfvars   # set domain_name, s3_bucket_name, region, etc.

# 3. Standard flow:
terraform init
terraform fmt -check
terraform validate
terraform plan -out tfplan
terraform apply tfplan
```

Key outputs after apply: `api_url`, `alb_dns_name`, `rds_endpoint`,
`redis_primary_endpoint`, `s3_bucket_name`, `ec2_instance_id`, `ssm_path_prefix`.

---

## How secrets flow (SSM → `.env`)

1. Terraform writes secrets as **SecureString** parameters under
   `var.ssm_path_prefix` (default `/edu-app/prod`). Parameter leaf names are
   `UPPER_SNAKE` and map 1:1 to `EDU_*` keys:
   - `EDU_JWT_SECRET`, `EDU_POSTGRES_URL` (with `sslmode=require`),
     `EDU_REDIS_URL` (`rediss://`), `POSTGRES_PASSWORD`, `EDU_SENTRY_DSN`,
     plus non-secret `EDU_S3_BUCKET` / `EDU_S3_REGION` / `EDU_S3_ENDPOINT` /
     `EDU_KAFKA_BROKERS`.
2. The EC2 `user_data` (`templates/user_data.sh.tftpl`) runs
   `aws ssm get-parameters-by-path --with-decryption --recursive` and renders
   `/opt/edu-app/.env` (mode `600`), then runs `deploy/scripts/deploy.sh`.
3. **No static S3 keys** are stored — the app/worker use the EC2 **instance
   role** for S3, and the bucket policy allows only that role.
4. **Rotation:** update the SSM value (or re-`apply` after changing a TF var),
   then on the host re-render `.env` and re-run `deploy.sh`. Rotating
   `EDU_JWT_SECRET` invalidates existing access tokens (do it in a low-traffic
   window) — see `../README.md` §3.

---

## How this maps to `deploy/scripts/deploy.sh`

- Terraform's `user_data` is a thin wrapper that prepares the host (Docker,
  repo at `/opt/edu-app`, `.env` from SSM) and then calls the **unchanged**
  `deploy/scripts/deploy.sh`, which builds/pulls images, runs the `migrate`
  job against RDS, brings up the compose prod stack, and smoke-tests.
- For subsequent releases you do **not** re-run Terraform. SSH/SSM into the host
  and run `deploy.sh` (optionally `IMAGE_TAG=... PULL=1 SKIP_BUILD=1` for a
  rollback), exactly as `../README.md` §5 describes.
- The ALB health check hits `/health`; `deploy.sh`'s smoke test still validates
  `/health` and `/health/ready` end-to-end over HTTPS.

---

## Kafka: MSK vs self-host

- **Phase 1 default (`enable_msk = false`):** no MSK is created. The compose
  stack keeps its KRaft `kafka` container and `EDU_KAFKA_BROKERS=kafka:29092`
  (the SSM param is set to that value).
- **`enable_msk = true`:** an MSK Serverless cluster with IAM auth is created and
  the instance role is granted `kafka-cluster:*` on it. Serverless bootstrap
  brokers are not known at apply time — fetch them and set the SSM param:

  ```sh
  BROKERS=$(aws kafka get-bootstrap-brokers \
    --cluster-arn "$(terraform output -raw ... )" \
    --query BootstrapBrokerStringSaslIam --output text)
  aws ssm put-parameter --overwrite --type String \
    --name /edu-app/prod/EDU_KAFKA_BROKERS --value "$BROKERS"
  ```

  The `EDU_KAFKA_BROKERS` param has `ignore_changes = [value]` so Terraform
  won't clobber the broker list you set out-of-band. Also disable the compose
  `kafka` container (`../docker-compose.prod.yml` MANAGED SERVICES stanza).

---

## Alternative: nginx on EC2 with an EIP (instead of the ALB)

The ALB is the **primary** approach (TLS at ACM, easy health checks, no public
IP on the box). If you prefer nginx terminating TLS on the instance:

1. Put the EC2 in a **public** subnet (`aws_subnet.public[0]`), attach an
   `aws_eip`, and open 80/443 to `0.0.0.0/0` on the app SG.
2. Drop the ALB, target group, and HTTPS/HTTP listeners; point the Route53 `A`
   record at the EIP instead of the ALB alias.
3. Use Let's Encrypt/certbot (or ACM-exported certs) at
   `deploy/nginx/certs/{fullchain,privkey}.pem`, as `../README.md` §1/§2 cover.

This trades the ALB's managed TLS + horizontal-scale path for a lower monthly
cost. Keep the ALB approach unless cost is the dominant constraint.

---

## Destroy / rollback

- **App rollback** (no infra change): on the host,
  `IMAGE_TAG=<prev> PULL=1 SKIP_BUILD=1 sh deploy/scripts/deploy.sh`. A rolled-
  back **image** does **not** roll back DB schema — restore the pre-deploy RDS
  snapshot if a migration was destructive (`../README.md` §4/§5).
- **Infra teardown:**

  ```sh
  terraform destroy
  ```

  Guardrails that will block a naive destroy (intentionally):
  - RDS has `deletion_protection = true` and `skip_final_snapshot = false`.
    Set `deletion_protection = false` (apply) and accept/keep the final
    snapshot before destroying, or you'll get an error.
  - The S3 bucket must be **emptied** (incl. all noncurrent versions) before it
    can be deleted; there is no `force_destroy` on the assets bucket by design.
  - Consider setting `enable_deletion_protection = true` on the ALB once live.

---

## Cost caveat

This is **production-shaped, not free**. Rough always-on cost drivers (region +
usage dependent, illustrative only):

- NAT Gateway (hourly + data) — the largest fixed cost; `single_nat_gateway`
  keeps it to one.
- ALB (hourly + LCU).
- EC2 `t3.large` + 50 GB gp3.
- RDS `db.t3.medium` (**doubles** with `db_multi_az = true`) + storage + backups.
- ElastiCache `cache.t3.micro`.
- MSK Serverless (only if `enable_msk = true`) — billed per partition-hour +
  throughput; **off by default** for Phase 1.

To trim early-stage cost: `single_nat_gateway = true` (default),
`db_multi_az = false`, smaller instance classes, and `enable_msk = false`. Turn
the resilience toggles back on before real production traffic.

---

## Notes / assumptions

- Region defaults to `ap-southeast-1` (matches the compose prod example).
- One EC2 instance (Phase 1). The target group + ALB make it straightforward to
  add more instances later.
- AMI is the latest Amazon Linux 2023 (looked up via SSM/EC2 data source).
- IMDSv2 is enforced; the instance has no public IP (egress via NAT).
- `terraform validate` could **not** be run in the authoring environment (no
  Terraform CLI / AWS creds available there). The HCL targets provider `~> 5.40`
  and uses only documented resources/arguments; run `terraform fmt` +
  `terraform validate` locally before applying.
