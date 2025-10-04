# Systemd Deployment for Kriti Images

This directory contains files and instructions for deploying Kriti Images as a systemd service on Linux servers.

## 📁 Files

- `kriti-images.service` - Systemd unit file
- `install.sh` - Automated installation script
- `README.md` - This documentation

## 🚀 Quick Installation

### Prerequisites

1. **Linux server** with systemd (Ubuntu 16.04+, CentOS 7+, Debian 9+, etc.)
2. **Go 1.24+** installed
3. **Root access** for installation
4. **System dependencies** (automatically installed by script)

### Automated Installation

```bash
# Clone the repository
git clone https://github.com/kritihq/kriti-images.git
cd kriti-images

# Run the installation script as root
sudo systemd/install.sh
```

The script will:
- Install system dependencies (libwebp, build tools)
- Create a dedicated service user
- Build the application with CGO support
- Install and configure the systemd service
- Start the service automatically

## 🔧 Manual Installation

If you prefer manual installation or need customization:

### Step 1: Install Dependencies

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install -y libwebp-dev build-essential wget curl
```

**RHEL/CentOS:**
```bash
sudo yum update -y
sudo yum groupinstall -y "Development Tools"
sudo yum install -y libwebp-devel wget curl
```

**Fedora:**
```bash
sudo dnf update -y
sudo dnf groupinstall -y "Development Tools"
sudo dnf install -y libwebp-devel wget curl
```

### Step 2: Create Service User

```bash
sudo useradd --system --no-create-home --shell /bin/false kriti-images
```

### Step 3: Create Directories

```bash
sudo mkdir -p /opt/kriti-images/web/static/assets
sudo mkdir -p /opt/kriti-images/web/templates
sudo mkdir -p /etc/kriti-images
sudo mkdir -p /var/log/kriti-images
```

### Step 4: Build Application

```bash
# Build with CGO enabled for WebP support
CGO_ENABLED=1 go build -o kriti-images main.go

# Copy files
sudo cp kriti-images /opt/kriti-images/
sudo cp -r web/* /opt/kriti-images/web/
sudo cp config.yaml /opt/kriti-images/
```

### Step 5: Set Permissions

```bash
sudo chown -R kriti-images:kriti-images /opt/kriti-images
sudo chown -R kriti-images:kriti-images /var/log/kriti-images
sudo chmod +x /opt/kriti-images/kriti-images
sudo chmod 775 /opt/kriti-images/web/static/assets
```

### Step 6: Install Systemd Service

```bash
sudo cp systemd/kriti-images.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable kriti-images
sudo systemctl start kriti-images
```

## 📋 Configuration

### Service Configuration

Edit `/opt/kriti-images/config.yaml`:

```yaml
server:
  port: 8080
  enable_print_routes: false
  read_timeout: 30s
  write_timeout: 30s

images:
  base_path: "/opt/kriti-images/web/static/assets"
  max_image_dimension: 8192
  max_file_size_in_bytes: 52428800

limiter:
  max: 100
  expiration: 1m
```

### Environment Variables

Create `/etc/kriti-images/environment` for additional environment variables:

```bash
# Environment variables for Kriti Images
GIN_MODE=release
LOG_LEVEL=info
```

### Adding Images

Copy your images to the assets directory:

```bash
sudo cp /path/to/your/images/* /opt/kriti-images/web/static/assets/
sudo chown -R kriti-images:kriti-images /opt/kriti-images/web/static/assets/
```

## 🔄 Service Management

### Basic Commands

```bash
# Start the service
sudo systemctl start kriti-images

# Stop the service
sudo systemctl stop kriti-images

# Restart the service
sudo systemctl restart kriti-images

# Check status
sudo systemctl status kriti-images

# Enable auto-start on boot
sudo systemctl enable kriti-images

# Disable auto-start
sudo systemctl disable kriti-images
```

### Monitoring

```bash
# View logs
sudo journalctl -u kriti-images

# Follow logs in real-time
sudo journalctl -u kriti-images -f

# View recent logs
sudo journalctl -u kriti-images --since "1 hour ago"

# View logs with priority
sudo journalctl -u kriti-images -p err
```

## 🌐 Testing the Installation

### Health Checks

```bash
# Check if service is responding
curl http://localhost:8080/health/ready

# Expected response: HTTP 200 OK

# Check liveness
curl http://localhost:8080/health/live
```

### Demo Page

Visit `http://your-server:8080/demo` to access the interactive demo.

### API Test

```bash
# Test image transformation (replace with your image)
curl "http://localhost:8080/cgi/images/tr:width=300,height=200/your-image.jpg"
```

## 🔒 Security Features

The systemd service includes several security hardening features:

- **Non-root execution**: Runs as dedicated `kriti-images` user
- **File system protection**: Read-only system directories
- **Capability restrictions**: Limited system capabilities
- **Namespace isolation**: Restricted access to system resources
- **Memory protection**: Write-execute memory prevention

### Firewall Configuration

If using UFW (Ubuntu):
```bash
sudo ufw allow 8080/tcp
```

If using firewalld (RHEL/CentOS):
```bash
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

## 🌍 Production Setup

### Reverse Proxy with Nginx

Create `/etc/nginx/sites-available/kriti-images`:

```nginx
server {
    listen 80;
    server_name your-domain.com;

    client_max_body_size 50M;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Caching for transformed images
        proxy_cache_valid 200 1h;
        proxy_cache_key "$scheme$request_method$host$request_uri";
    }
}
```

Enable the site:
```bash
sudo ln -s /etc/nginx/sites-available/kriti-images /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### SSL with Let's Encrypt

```bash
sudo apt-get install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

### Monitoring with Prometheus

The service exposes metrics at `/metrics`. Add to Prometheus config:

```yaml
scrape_configs:
  - job_name: 'kriti-images'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

## 🐛 Troubleshooting

### Common Issues

1. **Service won't start**
   ```bash
   # Check logs
   sudo journalctl -u kriti-images -n 50
   
   # Check configuration
   sudo -u kriti-images /opt/kriti-images/kriti-images --help
   ```

2. **Permission denied errors**
   ```bash
   # Fix ownership
   sudo chown -R kriti-images:kriti-images /opt/kriti-images
   sudo chown -R kriti-images:kriti-images /var/log/kriti-images
   ```

3. **Port already in use**
   ```bash
   # Check what's using port 8080
   sudo netstat -tlnp | grep 8080
   
   # Change port in config.yaml if needed
   ```

4. **WebP support issues**
   ```bash
   # Ensure libwebp is installed
   ldconfig -p | grep webp
   
   # Rebuild with CGO if needed
   CGO_ENABLED=1 go build -o kriti-images main.go
   ```

### Performance Tuning

1. **Increase file limits** in `/etc/systemd/system/kriti-images.service`:
   ```ini
   LimitNOFILE=65536
   LimitNPROC=4096
   ```

2. **Adjust Go runtime**:
   ```bash
   # Add to environment file
   GOMAXPROCS=4
   GOGC=100
   ```

3. **System-level optimizations**:
   ```bash
   # Increase kernel limits
   echo 'net.core.somaxconn = 1024' >> /etc/sysctl.conf
   sysctl -p
   ```

## 🔄 Updates and Maintenance

### Updating the Service

```bash
# Stop the service
sudo systemctl stop kriti-images

# Backup current installation
sudo cp -r /opt/kriti-images /opt/kriti-images.backup

# Build new version
CGO_ENABLED=1 go build -o kriti-images main.go

# Replace binary
sudo cp kriti-images /opt/kriti-images/

# Update web assets if changed
sudo cp -r web/* /opt/kriti-images/web/

# Restart service
sudo systemctl start kriti-images
```

### Log Rotation

Create `/etc/logrotate.d/kriti-images`:

```
/var/log/kriti-images/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 kriti-images kriti-images
    postrotate
        systemctl reload kriti-images
    endscript
}
```

### Backup Strategy

```bash
# Create backup script
cat > /usr/local/bin/backup-kriti-images.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/backup/kriti-images/$(date +%Y-%m-%d)"
mkdir -p "$BACKUP_DIR"
cp -r /opt/kriti-images "$BACKUP_DIR/"
cp -r /etc/kriti-images "$BACKUP_DIR/"
tar -czf "$BACKUP_DIR.tar.gz" "$BACKUP_DIR"
rm -rf "$BACKUP_DIR"
EOF

chmod +x /usr/local/bin/backup-kriti-images.sh

# Add to cron
echo "0 2 * * * /usr/local/bin/backup-kriti-images.sh" | sudo crontab -
```

## 📞 Support

- **Logs**: `sudo journalctl -u kriti-images -f`
- **Status**: `sudo systemctl status kriti-images`
- **Config**: `/opt/kriti-images/config.yaml`
- **Assets**: `/opt/kriti-images/web/static/assets`

For issues and support, check the main project repository.