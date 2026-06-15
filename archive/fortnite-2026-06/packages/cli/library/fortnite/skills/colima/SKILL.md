---
name: colima
description: "Docker runtime for macOS — start/stop/status/list instances, SSH, Kubernetes, resource management. Use when managing container runtime, Docker daemon, or local K8s clusters on macOS."
trigger: /colima
---

# /colima — Container Runtime for macOS

Colima provides container runtimes on macOS with minimal setup. Wraps Docker, containerd, and Kubernetes in a lightweight VM.

## When to Use

- Starting/stopping Docker runtime on macOS
- Checking Docker daemon status
- Managing multiple Colima instances (profiles)
- SSH into the Colima VM
- Managing local Kubernetes clusters
- Adjusting CPU/memory/disk resources
- Troubleshooting container runtime issues

## Quick Reference

### Basic Operations
```bash
colima start                    # Start Colima (default profile)
colima start --cpu 4 --memory 8 --disk 100  # Start with resources
colima stop                     # Stop Colima
colima restart                  # Restart Colima
colima status                   # Show current status
colima list                     # List all instances
colima delete                   # Delete and teardown Colima
```

### Resource Management
```bash
colima start --cpu 4            # 4 CPUs
colima start --memory 8         # 8 GB RAM
colima start --disk 100         # 100 GB disk
colima start --runtime docker   # Docker runtime (default)
colima start --runtime containerd  # containerd runtime
```

### Multiple Instances (Profiles)
```bash
colima start --profile dev      # Start 'dev' profile
colima start --profile prod     # Start 'prod' profile
colima status --profile dev     # Check 'dev' profile status
colima stop --profile dev       # Stop 'dev' profile
colima list                     # List all profiles
colima delete --profile dev     # Delete 'dev' profile
```

### SSH Access
```bash
colima ssh                      # SSH into VM
colima ssh-config               # Show SSH config
colima ssh --profile dev        # SSH into specific profile
```

### Kubernetes
```bash
colima kubernetes start         # Start K8s cluster
colima kubernetes stop          # Stop K8s cluster
colima kubernetes delete        # Delete K8s cluster
```

### Container Runtime
```bash
colima nerdctl -- version       # Run nerdctl (containerd runtime)
colima update                   # Update container runtime
```

### Maintenance
```bash
colima prune                    # Prune cached assets
colima template                 # Edit default config template
colima version                  # Print version
```

## Common Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-p, --profile` | `default` | Profile name for multiple instances |
| `--cpu` | 2 | Number of CPUs |
| `--memory` | 2 | Memory in GB |
| `--disk` | 60 | Disk size in GB |
| `--runtime` | `docker` | Container runtime (docker/containerd) |
| `--vm-type` | `vz` | VM type (vz/qemu) |

## Safety Rules

1. **Never run `colima delete` without human approval** — destroys all containers and data
2. **Check `colima status` before operations** — ensure runtime is running
3. **Use profiles for isolation** — separate dev/staging/prod environments
4. **Verify Docker socket path** — `~/.colima/default/docker.sock` for default profile
5. **Always stop before delete** — clean shutdown preferred

## Scripts

| Script | Purpose |
|--------|---------|
| `scripts/colima-health.sh` | Quick health check of Colima + Docker |
| `scripts/colima-resources.sh` | Show current resource allocation |

## Integration with Other Skills

- **dev-cli** — dev requires Docker runtime (colima provides it on macOS)
- **rift-deploy** — Uses colima for containerized deployments
- **respawn-crew** — Uses colima health checks during incident response

## Agent Ownership

Primary: **rift-deploy** (container runtime management)
Secondary: **respawn-crew** (incident health checks), **wall-builder** (dev testing prerequisites)
