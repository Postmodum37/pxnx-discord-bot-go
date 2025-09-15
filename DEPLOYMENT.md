# Docker Deployment Guide

This guide explains how to deploy the PXNX Discord Bot using Docker and Docker Compose on your remote server.

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- Git (for cloning the repository)

## Quick Start

### 1. Clone the Repository
```bash
git clone https://github.com/Postmodum37/pxnx-discord-bot-go.git
cd pxnx-discord-bot-go
```

### 2. Configure Environment
```bash
# Copy the example environment file
cp .env.example .env

# Edit the environment file with your tokens
nano .env
```

Required environment variables:
```env
DISCORD_BOT_TOKEN=your_discord_bot_token_here
OPENWEATHER_API_KEY=your_openweather_api_key_here
```

### 3. Deploy with Docker Compose
```bash
# Pull the latest image and start the bot
docker-compose up -d

# View logs
docker-compose logs -f pxnx-discord-bot

# Check status
docker-compose ps
```

### 4. Register Discord Commands (First Time Only)
```bash
# Register slash commands with Discord
docker-compose exec pxnx-discord-bot ./pxnx-discord-bot --register-commands

# Or with temporary container
docker run --rm --env-file .env ghcr.io/postmodum37/pxnx-discord-bot-go:latest ./pxnx-discord-bot --register-commands
```

## Management Commands

### Start/Stop the Bot
```bash
# Start
docker-compose up -d

# Stop
docker-compose down

# Restart
docker-compose restart
```

### View Logs
```bash
# Follow logs in real-time
docker-compose logs -f

# View last 100 lines
docker-compose logs --tail=100

# View logs for specific time
docker-compose logs --since="2h"
```

### Update to Latest Version
```bash
# Pull latest image
docker-compose pull

# Recreate containers with new image
docker-compose up -d
```

## Advanced Configuration

### Custom Docker Compose Override
Create `docker-compose.override.yml` for custom settings:

```yaml
version: '3.8'
services:
  pxnx-discord-bot:
    # Custom resource limits
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '1.0'

    # Custom restart policy
    restart: always

    # Mount volume for logs
    volumes:
      - ./logs:/app/logs
```

### Build from Source (Optional)
If you prefer to build locally instead of using pre-built images:

```yaml
# In docker-compose.yml, replace the image line with:
build:
  context: .
  dockerfile: Dockerfile
```

Then run:
```bash
docker-compose build --no-cache
docker-compose up -d
```

### Health Monitoring
The container includes health checks. Monitor with:

```bash
# Check health status
docker-compose ps

# View detailed health check logs
docker inspect $(docker-compose ps -q pxnx-discord-bot) | grep -A 10 "Health"
```

## Security Best Practices

### 1. Environment File Security
```bash
# Set proper permissions on .env file
chmod 600 .env

# Ensure .env is in .gitignore
echo ".env" >> .gitignore
```

### 2. Container Security
The container runs as a non-root user (uid 1001) and includes:
- Read-only root filesystem capability
- No new privileges security option
- Minimal Alpine Linux base image
- Regular security scanning via Trivy

### 3. Network Security
```bash
# Optional: Create custom network for isolation
docker network create bot-network

# Update docker-compose.yml to use custom network
```

## Troubleshooting

### Common Issues

1. **Bot not responding to commands**
   ```bash
   # Check if commands are registered
   docker-compose logs | grep "command"

   # Re-register commands
   docker-compose exec pxnx-discord-bot ./pxnx-discord-bot --register-commands
   ```

2. **Container keeps restarting**
   ```bash
   # Check logs for errors
   docker-compose logs pxnx-discord-bot

   # Verify environment variables
   docker-compose exec pxnx-discord-bot env | grep -E "(DISCORD|WEATHER)"
   ```

3. **Permission issues**
   ```bash
   # Fix file permissions
   sudo chown -R $USER:$USER .
   chmod 600 .env
   ```

4. **Out of memory**
   ```bash
   # Check container resource usage
   docker stats pxnx-discord-bot

   # Increase memory limit in docker-compose.yml
   ```

### Debug Mode
Run container interactively for debugging:
```bash
docker run -it --env-file .env ghcr.io/postmodum37/pxnx-discord-bot-go:latest sh
```

## Monitoring and Maintenance

### Log Rotation
Docker Compose is configured with log rotation:
- Max file size: 10MB
- Max files: 3
- Total max logs: ~30MB

### System Resources
Recommended minimum system requirements:
- RAM: 512MB available
- CPU: 1 core
- Disk: 1GB free space

### Backup
Backup your configuration:
```bash
# Backup environment file
cp .env .env.backup

# Backup entire configuration
tar -czf bot-backup-$(date +%Y%m%d).tar.gz .env docker-compose.yml
```

## CI/CD Integration

The Docker image is automatically built and pushed to GitHub Container Registry on every commit to the main branch. Available tags:

- `latest` - Latest commit from main branch
- `main-<sha>` - Specific commit SHA
- `v*` - Release tags (if using)

## Support

If you encounter issues:

1. Check the [troubleshooting section](#troubleshooting) above
2. Review container logs: `docker-compose logs -f`
3. Open an issue on [GitHub](https://github.com/Postmodum37/pxnx-discord-bot-go/issues)

## Performance Tuning

For high-load scenarios:
```yaml
# In docker-compose.yml
deploy:
  resources:
    limits:
      memory: 512M
      cpus: '2.0'
    reservations:
      memory: 128M
      cpus: '0.5'
```