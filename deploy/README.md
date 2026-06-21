# Edu App — Production Deployment Runbook (Workstream C)

Infra-as-runbook for deploying the Edu App stack (Go API + Python worker + admin)
to AWS: EC2 behind Nginx+TLS, backed by RDS / ElastiCache / MSK / S3.

This document covers **Workstream C** of `plan/PHASE1_GOLIVE_PLAN.md` (§7). Its
exit criteria are:

- Domain + HTTPS live
- API healthcheck passes on the real domain
- Backup + restore test passes at least once

> **No real secrets live in this repo.** Everything below uses placeholders.
> Inject secrets at runtime from SSM Parameter Store / Secrets Manager into a
> host-side `.env` (gitignored) or the process environment.

---

## Files in this directory

| File | Purpose |
|------|---------|
| `docker-compose.prod.yml` | Production override (nginx + TLS, internal-only data services, env-driven secrets). |
| `nginx/edu-app.conf` | Reverse proxy, HTTP→HTTPS, security headers, `/metrics` lockdown. |
| `nginx/certs/` | Mount point for TLS certs (`fullchain.pem`, `privkey.pem`) — not committed. |
| `scripts/deploy.sh` | Build/pull → migrate → up → smoke test. Idempotent. |
| `scripts/smoke-test.sh` | Post-deploy health verification. |

Run the stack from the **repo root**:

```sh
docker compose -f docker-compose.yml -f deploy/docker-compose.prod.yml up -d
```

---

## 1. AWS provisioning checklist (C1)

### Networking
- [ ] **VPC** (e.g. `10.0.0.0/16`).
- [ ] **Subnets**: ≥2 public (ALB/NAT/EC2) + ≥2 private (RDS/ElastiCache/MSK) across 2 AZs.
- [ ] **Internet Gateway** on the VPC; **NAT Gateway** for private-subnet egress.
- [ ] **Route tables**: public → IGW, private → NAT.

### Security groups (least privilege)
- [ ] `sg-edge` (EC2): inbound 443 (and 80 for redirect) from `0.0.0.0/0`; SSH 22 from your admin IP only.
- [ ] `sg-rds`: inbound 5432 from `sg-edge` only.
- [ ] `sg-redis`: inbound 6379 from `sg-edge` only.
- [ ] `sg-msk`: inbound 9098 (IAM auth) / 9092 from `sg-edge` only.
- [ ] No data-store SG opens to `0.0.0.0/0`.

### IAM
- [ ] **EC2 instance role** with: `s3:GetObject/PutObject/ListBucket` on the asset bucket,
      `ssm:GetParameter*` (or `secretsmanager:GetSecretValue`) for the app's secrets path,
      and CloudWatch Logs write. Prefer the instance role over static S3 keys.
- [ ] **CI deploy role** (if using OIDC for ECR push) with ECR push permissions.

### Compute
- [ ] **EC2**: start at `t3.large` (2 vCPU / 8 GB) for the full compose stack; scale to
      `m5.large`+ under load. Bump if you keep Postgres/Kafka on-box. EBS gp3 ≥50 GB.
- [ ] Install Docker + Compose v2. Clone the repo to `/opt/edu-app`.

### Managed data services
- [ ] **RDS PostgreSQL 16**, Multi-AZ, in private subnets. Connection string **must**
      use `sslmode=require` (the app refuses to boot in production otherwise —
      see `server/config/config.go` `validateProduction`).
- [ ] **ElastiCache Redis 7** (cluster or single-node) in private subnets; enable
      in-transit TLS and use a `rediss://` URL.
- [ ] **MSK Serverless** (recommended) or self-host Kafka. With MSK use IAM auth on
      port 9098 and set `EDU_KAFKA_BROKERS` to the bootstrap brokers. *Self-host note:*
      keeping the compose `kafka` (KRaft) container is acceptable for low volume —
      just leave that container enabled and point `EDU_KAFKA_BROKERS=kafka:29092`.
- [ ] **S3 bucket** (`edu-assets-prod`), **Block Public Access = ON**, default
      encryption (SSE-S3/KMS), versioning on. Bucket policy denies public access;
      access is via the EC2 instance role only.

### DNS + TLS
- [ ] **Route53** A/ALIAS record for the API domain → EC2 (or an ALB).
- [ ] **TLS**: either terminate at an **ALB with ACM** (then nginx can be HTTP-only
      behind it), or terminate at nginx with **Let's Encrypt** certs on the box.
      This runbook assumes nginx terminates TLS using certs in `nginx/certs/`.

### Sample private S3 bucket policy (placeholder ARNs)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "DenyInsecureTransport",
      "Effect": "Deny",
      "Principal": "*",
      "Action": "s3:*",
      "Resource": [
        "arn:aws:s3:::edu-assets-prod",
        "arn:aws:s3:::edu-assets-prod/*"
      ],
      "Condition": { "Bool": { "aws:SecureTransport": "false" } }
    },
    {
      "Sid": "AllowAppRole",
      "Effect": "Allow",
      "Principal": { "AWS": "arn:aws:iam::<ACCOUNT_ID>:role/edu-app-ec2-role" },
      "Action": ["s3:GetObject", "s3:PutObject", "s3:ListBucket"],
      "Resource": [
        "arn:aws:s3:::edu-assets-prod",
        "arn:aws:s3:::edu-assets-prod/*"
      ]
    }
  ]
}
```

---

## 2. Server provisioning (C2)

On the EC2 host:

1. Install Docker Engine + Compose plugin.
2. Clone the repo to `/opt/edu-app` (matches `REMOTE_REPO_DIR` in `deploy.yml`).
3. Place TLS certs at `deploy/nginx/certs/fullchain.pem` + `privkey.pem`
   (or set `EDU_TLS_CERT_DIR` to wherever certbot/ACM exports them).
4. Create `/opt/edu-app/.env` from secrets (see §3). It is gitignored.
5. Edit `deploy/nginx/edu-app.conf` `server_name` to your real domain, and the
   `/metrics` `allow` CIDR to your VPC/monitoring range.

### Log rotation

Configure the Docker daemon `/etc/docker/daemon.json` so container logs rotate:

```json
{
  "log-driver": "json-file",
  "log-opts": { "max-size": "20m", "max-file": "5" }
}
```

Restart Docker after changing this. (Optionally ship logs to CloudWatch via the
awslogs driver / CloudWatch agent.) Nginx access/error logs go to the container's
stdout/stderr and inherit the same rotation.

---

## 3. Secrets management (C3)

**Never commit secrets.** Source them from SSM Parameter Store (SecureString) or
Secrets Manager and render a host-side `.env`:

```sh
# Example: pull a parameter tree into .env (placeholders).
aws ssm get-parameters-by-path --with-decryption --path /edu-app/prod \
  --query "Parameters[].{Name:Name,Value:Value}" --output text \
  | awk '{ n=$1; sub(".*/","",n); print toupper(n) "=" $2 }' > /opt/edu-app/.env
```

Required keys (consumed by `deploy/docker-compose.prod.yml`):

| Key | Notes |
|-----|-------|
| `EDU_JWT_SECRET` | ≥32 chars, non-default. `openssl rand -base64 48`. |
| `EDU_POSTGRES_URL` | RDS URL with `sslmode=require`. |
| `EDU_REDIS_URL` | ElastiCache `rediss://…` (TLS). |
| `EDU_KAFKA_BROKERS` | MSK bootstrap brokers, or `kafka:29092` self-host. |
| `EDU_S3_ACCESS_KEY` / `EDU_S3_SECRET_KEY` | IAM creds (or omit + use instance role). |
| `EDU_S3_BUCKET`, `EDU_S3_REGION`, `EDU_S3_ENDPOINT` | empty endpoint = real AWS S3. |
| `POSTGRES_PASSWORD` | only if running the on-box postgres container. |
| `EDU_SENTRY_DSN` | optional; empty disables Sentry. |
| `EDU_TLS_CERT_DIR` | host dir holding the TLS certs. |

**Rotation policy:** rotate `EDU_JWT_SECRET` and DB credentials on a schedule.
Rotating the JWT secret invalidates all existing access tokens (clients re-login);
do it during a low-traffic window. Update the SSM/Secrets value, re-render `.env`,
then re-run `deploy.sh`.

### Pointing at managed services

`docker-compose.prod.yml` defaults the data URLs to the in-network containers but
lets env override them. When you move to RDS/ElastiCache/MSK/S3, set the URLs in
`.env` **and disable the now-redundant local containers** by uncommenting the
`deploy.replicas: 0` stanzas in `docker-compose.prod.yml` (one per service). For
real S3, disable both `minio` and `minio-init` and leave `EDU_S3_ENDPOINT` empty.

---

## 4. Data layer readiness (C4)

### Migrations
`deploy.sh` runs the one-shot `migrate` job (`/app/migrate up`) against RDS before
starting the app. It is idempotent — re-running is a no-op when current.

### Backup snapshot policy
- [ ] RDS **automated backups** enabled, retention ≥7 days; daily snapshot window set.
- [ ] On-demand snapshot before each production deploy/migration.
- [ ] S3 **versioning** on; add a lifecycle rule to expire noncurrent versions after N days.

### Restore test (exit criterion — must pass ≥1×)
1. Restore the latest RDS snapshot to a **temporary** instance
   (`aws rds restore-db-instance-from-db-snapshot`).
2. Point a throwaway `EDU_POSTGRES_URL` at it (keep `sslmode=require`).
3. Run `migrate` against it (should report up-to-date) and run the smoke test
   with `RUN_AUTH_TEST=1`.
4. Verify row counts / a known record. Tear down the temp instance.
5. Record the date + result. This satisfies the "backup + restore test pass" criterion.

---

## 5. Deploy & rollback (C5)

### Deploy

```sh
cd /opt/edu-app
# secrets in .env; certs in deploy/nginx/certs
sh deploy/scripts/deploy.sh
```

`deploy.sh` builds/pulls images, runs migrations, brings the stack up with the
prod override, then runs `smoke-test.sh`. A non-200 health check fails the deploy.

CI path: `.github/workflows/deploy.yml` (scaffold) builds + pushes the backend
image, then SSHes to EC2 and runs `deploy.sh` with `PULL=1 SKIP_BUILD=1`.

### Smoke test (manual)

```sh
# On the box (talks to nginx over TLS + app:8080 internally):
BASE_URL=https://api.example.com RUN_AUTH_TEST=1 sh deploy/scripts/smoke-test.sh
```

Password policy for the auth round-trip: **10–72 chars, must include a letter and a digit.**

### Rollback (re-deploy previous image tag)

Images are tagged per release. To roll back, redeploy the prior tag — no rebuild:

```sh
cd /opt/edu-app
IMAGE_TAG=v1.2.3 PULL=1 SKIP_BUILD=1 sh deploy/scripts/deploy.sh
```

> **Migration caution:** rolling the *image* back does not roll back DB schema.
> If a deploy applied a destructive migration, restore from the pre-deploy RDS
> snapshot (§4) instead of just reverting the image.

---

## 6. Workstream C exit-criteria checklist

- [ ] Domain resolves; HTTPS serves a valid cert (HTTP redirects to HTTPS).
- [ ] `https://<domain>/health` and `/health/ready` return 200 (via `smoke-test.sh`).
- [ ] `/metrics` is reachable internally but **denied** publicly (nginx allow-list).
- [ ] RDS migration passes; backup + restore test completed and recorded ≥1×.
- [ ] Secrets sourced from SSM/Secrets Manager; none committed to the repo.
- [ ] Log rotation configured; `restart: unless-stopped` on app/worker/nginx.
