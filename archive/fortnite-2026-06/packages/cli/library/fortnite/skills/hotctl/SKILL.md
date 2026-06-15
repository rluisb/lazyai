---
name: hotctl
description: "Hotmart infrastructure CLI — AWS ECR, EKS, RDS, SQS, Secrets, Terraform, SSO, Shield. Use when working with AWS infrastructure, container registries, Kubernetes clusters, databases, message queues, or CI/CD pipeline migration."
trigger: /hotctl
---

# /hotctl — Hotmart Infrastructure CLI

Hotmart's internal CLI for managing AWS infrastructure. Wraps common AWS operations with company-specific profiles, roles, and conventions.

## When to Use

- Working with AWS ECR (container registries)
- Managing EKS clusters (Kubernetes contexts)
- RDS database operations (tokens, migrations, validation)
- SQS message queue transfers
- AWS Secrets Manager operations
- Terraform infrastructure management
- AWS SSO login and account management
- CI/CD migration (Drone → GitHub Actions)
- AWS Shield incident response

## Quick Reference

### Authentication
```bash
hotctl sso init          # Initialize SSO config
hotctl sso login         # Login to AWS SSO
hotctl sso accounts      # List/manage AWS accounts
```

### Container Registry (ECR)
```bash
hotctl ecr login                    # Login to ECR (default profile/region)
hotctl ecr login --profile prod     # Login with specific profile
hotctl ecr login --region us-east-1 # Login with specific region
```

### Kubernetes (EKS)
```bash
hotctl eks list                     # List all EKS clusters across accounts
hotctl eks context --profile prod   # Switch kube context to profile
hotctl eks show                     # Show current EKS context
```

### Database (RDS)
```bash
hotctl rds list                     # List all RDS clusters
hotctl rds login --cluster NAME     # Login to RDS cluster
hotctl rds generate --cluster NAME  # Generate DB auth token
hotctl rds validate                 # Validate database health
hotctl rds migrate                  # Run database migrations
hotctl rds create                   # Create RDS resources
```

### Message Queues (SQS)
```bash
hotctl sqs transfer --from URL --to URL  # Transfer messages between queues
```

### Secrets Manager
```bash
hotctl secret list                  # List all secrets
hotctl secret get --name NAME       # Get specific secret value
hotctl secret export --name NAME    # Export secret as bash env var
hotctl secret fixrotate --name NAME # Fix failed secret rotation
```

### Infrastructure (Terraform)
```bash
hotctl infra delete --state S3_URL  # Delete infra based on S3 terraform state
```

### Git Operations
```bash
hotctl git merge                    # Merge iac-wrapper commands
```

### CI/CD Migration
```bash
hotctl drone2gha .drone.yml --namespace myns                    # Convert to multiple pipelines
hotctl drone2gha .drone.yml --single-pipeline --namespace myns  # Convert to single pipeline
hotctl drone2gha .drone.yml --output workflow.yml --namespace myns --single-pipeline
```

### AWS Shield
```bash
hotctl shield engage    # Open case for AWS Shield Response Team
```

### Repository Info
```bash
hotctl explain          # Show info about Hotmart repositories
```

## Common Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--profile` | `default` | AWS profile for specific account |
| `--region` | `us-east-1` | AWS region |
| `-r, --role` | (none) | AWS profile role |
| `--silent` | false | Omit log output |

## Safety Rules

1. **Never run `hotctl infra delete` without human approval** — destructive operation
2. **Never run `hotctl rds migrate` without human approval** — database changes
3. **Always use `--silent` in automated scripts** — cleaner output for parsing
4. **Verify SSO login before any AWS operation** — run `hotctl sso login` first
5. **Use `hotctl eks show` to confirm context** before kubectl operations
6. **Never export secrets to logs** — use `hotctl secret export` only in secure contexts

## Scripts

| Script | Purpose |
|--------|---------|
| `scripts/hotctl-status.sh` | Quick health check of all AWS services |
| `scripts/hotctl-context.sh` | Show current AWS context (profile, region, EKS, ECR) |

## Integration with Other Skills

- **rift-deploy** — Uses hotctl for AWS infrastructure operations during deployment
- **respawn-crew** — Uses hotctl for incident response (Shield, RDS, SQS)
- **wall-builder** — Uses hotctl for infra-related code changes (ECR login for builds)

## Agent Ownership

Primary: **wall-builder** (infra ops), **rift-deploy** (deploy infra)
Secondary: **respawn-crew** (incident response)
