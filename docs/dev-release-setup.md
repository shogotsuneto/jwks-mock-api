# Dev Release Pipeline Setup

This document provides step-by-step instructions for setting up and using the dev release pipeline.

## Automatic Setup (Minimal Requirements)

The dev release pipeline uses GitHub's built-in `GITHUB_TOKEN` and requires **no additional manual configuration** for basic functionality.

### What Happens Automatically

1. **On Push to `develop` Branch**: Pipeline triggers automatically
2. **Binary Build**: Creates optimized x86_64 Linux binary
3. **Artifact Upload**: Binary uploaded to GitHub Actions artifacts (30-day retention)
4. **Container Build**: Multi-tagged Docker image built and pushed to GHCR
5. **Registry Access**: Uses `GITHUB_TOKEN` with automatic package permissions

## Quick Start

### Access Published Artifacts

**GitHub Artifacts** (Web UI):
- Go to repository → Actions → Select workflow run
- Download artifacts from the "Artifacts" section

**Container Images** (CLI):
```bash
# Pull latest dev image
docker pull ghcr.io/shogotsuneto/jwks-mock-api:develop-latest

# Pull specific commit
docker pull ghcr.io/shogotsuneto/jwks-mock-api:develop-<sha>

# Run container
docker run -p 3000:3000 ghcr.io/shogotsuneto/jwks-mock-api:develop-latest
```

## Advanced Configuration (Optional)

### Custom Registry Configuration

If you need to use a different container registry:

1. **Add Registry Secrets**:
   - Go to repository Settings → Secrets and variables → Actions
   - Add secrets for your registry (e.g., `CUSTOM_REGISTRY`, `CUSTOM_TOKEN`)

2. **Modify Workflow**:
   - Edit `.github/workflows/dev-release.yml`
   - Update `REGISTRY` environment variable
   - Update login action with your registry credentials

### Private Repository Considerations

- **GHCR Images**: Automatically inherit repository visibility
- **Access Control**: Use GitHub's package permissions for fine-grained access
- **Pull Access**: Users need appropriate repository permissions to pull images

## Troubleshooting

### Pipeline Failures

**Common Issues**:
1. **Build failures**: Check Go code builds locally with `make build`
2. **Docker build issues**: Test locally with `make docker`

**Debug Steps**:
```bash
# Test locally first
make clean && make build-optimized
make docker
docker run --rm jwks-mock-api:latest --help
```

### Registry Access Issues

**GHCR Authentication**:
- Pipeline uses `GITHUB_TOKEN` automatically
- For manual access: `echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin`
- For personal access: Use GitHub Personal Access Token with `read:packages` scope

## Pipeline Output

Each successful run produces:

**Build Artifacts**:
- `jwks-mock-api-develop-<sha>-<timestamp>` binary (GitHub artifacts)

**Container Images**:
- `ghcr.io/shogotsuneto/jwks-mock-api:develop-latest`
- `ghcr.io/shogotsuneto/jwks-mock-api:develop-<sha>`
- `ghcr.io/shogotsuneto/jwks-mock-api:develop-<sha>-<timestamp>`

**Metadata**:
- Build summary in GitHub Actions UI
- Image labels with commit info and build date