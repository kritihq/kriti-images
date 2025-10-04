# Docker Setup for Kriti Images

This directory contains Docker-related files for building and running the Kriti Images service.

## Files

- `Dockerfile` - Multi-stage Docker build configuration
- `.dockerignore` - Files to exclude from Docker build context

## Quick Start

### Building the Image

From the project root directory:

```bash
# Basic build
docker build -t kriti-images -f docker/Dockerfile .
```

### Running the Container

```bash
# Basic run (ephemeral images)
docker run -p 8080:8080 kriti-images

# With persistent image storage
docker run -p 8080:8080 -v /path/to/your/images:/app/web/static/assets kriti-images

# With custom config
docker run -p 8080:8080 \
  -v /path/to/your/images:/app/web/static/assets \
  -v /path/to/config.yaml:/app/config.yaml \
  kriti-images

# Run in background
docker run -d -p 8080:8080 \
  -v /path/to/your/images:/app/web/static/assets \
  --name kriti-images-server \
  kriti-images
```
