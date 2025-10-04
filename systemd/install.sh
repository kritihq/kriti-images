#!/bin/bash

# Kriti Images - Systemd Installation Script
# This script installs and configures Kriti Images as a systemd service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
INSTALL_DIR="/opt/kriti-images"
SERVICE_USER="kriti-images"
SERVICE_GROUP="kriti-images"
CONFIG_DIR="/etc/kriti-images"
LOG_DIR="/var/log/kriti-images"
BINARY_NAME="kriti-images"
SERVICE_FILE="kriti-images.service"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root"
        exit 1
    fi
}

# Function to check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.24+ first."
        print_status "Visit: https://golang.org/doc/install"
        exit 1
    fi

    GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
    print_status "Found Go version: $GO_VERSION"
}

# Function to install system dependencies
install_dependencies() {
    print_status "Installing system dependencies..."

    # Detect package manager and install dependencies
    if command -v apt-get &> /dev/null; then
        # Debian/Ubuntu
        apt-get update
        apt-get install -y wget curl libwebp-dev build-essential
    elif command -v yum &> /dev/null; then
        # RHEL/CentOS
        yum update -y
        yum groupinstall -y "Development Tools"
        yum install -y wget curl libwebp-devel
    elif command -v dnf &> /dev/null; then
        # Fedora
        dnf update -y
        dnf groupinstall -y "Development Tools"
        dnf install -y wget curl libwebp-devel
    elif command -v pacman &> /dev/null; then
        # Arch Linux
        pacman -Sy --noconfirm wget curl libwebp base-devel
    elif command -v apk &> /dev/null; then
        # Alpine Linux
        apk update
        apk add --no-cache wget curl libwebp-dev build-base
    else
        print_warning "Unknown package manager. Please install libwebp development libraries manually."
    fi

    print_success "System dependencies installed"
}

# Function to create service user
create_user() {
    print_status "Creating service user: $SERVICE_USER"

    if id "$SERVICE_USER" &>/dev/null; then
        print_warning "User $SERVICE_USER already exists"
    else
        useradd --system --no-create-home --shell /bin/false "$SERVICE_USER"
        print_success "Created user: $SERVICE_USER"
    fi
}

# Function to create directories
create_directories() {
    print_status "Creating directories..."

    # Create installation directory
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$INSTALL_DIR/web/static/assets"
    mkdir -p "$INSTALL_DIR/web/templates"

    # Create config directory
    mkdir -p "$CONFIG_DIR"

    # Create log directory
    mkdir -p "$LOG_DIR"

    print_success "Directories created"
}

# Function to build the application
build_application() {
    print_status "Building Kriti Images application..."

    # Get the script directory (where the source code should be)
    SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
    SOURCE_DIR="$(dirname "$SCRIPT_DIR")"

    if [[ ! -f "$SOURCE_DIR/main.go" ]]; then
        print_error "main.go not found in $SOURCE_DIR"
        print_error "Please run this script from the systemd directory of the project"
        exit 1
    fi

    cd "$SOURCE_DIR"

    # Build with CGO enabled for webp support
    print_status "Compiling Go application..."
    CGO_ENABLED=1 go build -o "$BINARY_NAME" main.go

    if [[ ! -f "$BINARY_NAME" ]]; then
        print_error "Build failed - binary not found"
        exit 1
    fi

    print_success "Application built successfully"

    # Copy binary to installation directory
    cp "$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"

    # Copy web assets
    cp -r web/* "$INSTALL_DIR/web/"

    # Copy config file
    if [[ -f "config.yaml" ]]; then
        cp config.yaml "$INSTALL_DIR/"
    else
        print_warning "config.yaml not found, you'll need to create one"
    fi

    print_success "Application files copied to $INSTALL_DIR"
}

# Function to set permissions
set_permissions() {
    print_status "Setting file permissions..."

    # Set ownership
    chown -R "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR"
    chown -R "$SERVICE_USER:$SERVICE_GROUP" "$LOG_DIR"

    # Set permissions
    chmod 755 "$INSTALL_DIR"
    chmod 755 "$INSTALL_DIR/$BINARY_NAME"
    chmod -R 755 "$INSTALL_DIR/web"
    chmod 775 "$INSTALL_DIR/web/static/assets"  # Allow writing for uploaded images

    print_success "Permissions set"
}

# Function to install systemd service
install_service() {
    print_status "Installing systemd service..."

    SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

    if [[ ! -f "$SCRIPT_DIR/$SERVICE_FILE" ]]; then
        print_error "Service file $SERVICE_FILE not found in $SCRIPT_DIR"
        exit 1
    fi

    # Copy service file
    cp "$SCRIPT_DIR/$SERVICE_FILE" "/etc/systemd/system/"

    # Reload systemd
    systemctl daemon-reload

    print_success "Systemd service installed"
}

# Function to create environment file
create_environment() {
    print_status "Creating environment configuration..."

    cat > "$CONFIG_DIR/environment" << EOF
# Kriti Images Environment Configuration
# Add any environment variables here

# Uncomment and modify as needed:
# GIN_MODE=release
# LOG_LEVEL=info
EOF

    chmod 640 "$CONFIG_DIR/environment"
    chown root:"$SERVICE_GROUP" "$CONFIG_DIR/environment"

    print_success "Environment file created at $CONFIG_DIR/environment"
}

# Function to create sample config
create_sample_config() {
    if [[ ! -f "$INSTALL_DIR/config.yaml" ]]; then
        print_status "Creating sample configuration..."

        cat > "$INSTALL_DIR/config.yaml" << EOF
server:
  port: 8080
  enable_print_routes: false
  read_timeout: 30s
  write_timeout: 30s

images:
  base_path: "/opt/kriti-images/web/static/assets"
  max_image_dimension: 8192 # 8k
  max_file_size_in_bytes: 52428800 # 50MB

limiter:
  max: 100
  expiration: 1m
EOF

        chown "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR/config.yaml"
        print_success "Sample configuration created"
    fi
}

# Function to enable and start service
start_service() {
    print_status "Enabling and starting Kriti Images service..."

    # Enable service to start on boot
    systemctl enable "$SERVICE_FILE"

    # Start the service
    systemctl start "$SERVICE_FILE"

    # Wait a moment and check status
    sleep 2

    if systemctl is-active --quiet "$SERVICE_FILE"; then
        print_success "Kriti Images service is running!"
        systemctl status "$SERVICE_FILE" --no-pager -l
    else
        print_error "Failed to start Kriti Images service"
        print_status "Check logs with: journalctl -u $SERVICE_FILE -f"
        exit 1
    fi
}

# Function to show post-installation info
show_info() {
    echo
    print_success "Installation completed successfully!"
    echo
    echo -e "${BLUE}Service Management:${NC}"
    echo "  Start:   systemctl start $SERVICE_FILE"
    echo "  Stop:    systemctl stop $SERVICE_FILE"
    echo "  Restart: systemctl restart $SERVICE_FILE"
    echo "  Status:  systemctl status $SERVICE_FILE"
    echo "  Logs:    journalctl -u $SERVICE_FILE -f"
    echo
    echo -e "${BLUE}Configuration:${NC}"
    echo "  Service config: $INSTALL_DIR/config.yaml"
    echo "  Environment:    $CONFIG_DIR/environment"
    echo "  Images folder:  $INSTALL_DIR/web/static/assets"
    echo
    echo -e "${BLUE}Service Access:${NC}"
    echo "  Health check: curl http://localhost:8080/health/ready"
    echo "  Demo page:    http://localhost:8080/demo"
    echo "  Transform API: http://localhost:8080/cgi/images/tr:<params>/<image>"
    echo
    print_warning "Don't forget to:"
    echo "  1. Add your images to $INSTALL_DIR/web/static/assets"
    echo "  2. Adjust firewall rules if needed (port 8080)"
    echo "  3. Configure reverse proxy (nginx/apache) for production"
}

# Main installation function
main() {
    echo -e "${GREEN}Kriti Images - Systemd Installation Script${NC}"
    echo "=============================================="
    echo

    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --install-dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            --user)
                SERVICE_USER="$2"
                SERVICE_GROUP="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --install-dir DIR    Installation directory (default: /opt/kriti-images)"
                echo "  --user USER          Service user (default: kriti-images)"
                echo "  --help               Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done

    print_status "Starting installation with:"
    echo "  Install directory: $INSTALL_DIR"
    echo "  Service user: $SERVICE_USER"
    echo

    # Run installation steps
    check_root
    check_go
    install_dependencies
    create_user
    create_directories
    build_application
    set_permissions
    install_service
    create_environment
    create_sample_config
    start_service
    show_info
}

# Run main function
main "$@"
