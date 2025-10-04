#!/bin/bash

# Kriti Images - Systemd Uninstallation Script
# This script removes Kriti Images systemd service and all related files

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

# Function to confirm uninstallation
confirm_uninstall() {
    echo -e "${YELLOW}WARNING: This will completely remove Kriti Images and all its data!${NC}"
    echo
    echo "This will remove:"
    echo "  - Systemd service: /etc/systemd/system/$SERVICE_FILE"
    echo "  - Installation directory: $INSTALL_DIR"
    echo "  - Configuration directory: $CONFIG_DIR"
    echo "  - Log directory: $LOG_DIR"
    echo "  - Service user: $SERVICE_USER"
    echo

    if [[ "$FORCE_UNINSTALL" != "true" ]]; then
        read -p "Are you sure you want to continue? (yes/no): " -r
        if [[ ! $REPLY =~ ^[Yy]es$ ]]; then
            print_status "Uninstallation cancelled"
            exit 0
        fi
    fi

    echo
    print_status "Proceeding with uninstallation..."
}

# Function to stop and disable service
stop_service() {
    print_status "Stopping and disabling Kriti Images service..."

    if systemctl is-active --quiet "$SERVICE_FILE"; then
        systemctl stop "$SERVICE_FILE"
        print_success "Service stopped"
    else
        print_warning "Service was not running"
    fi

    if systemctl is-enabled --quiet "$SERVICE_FILE" 2>/dev/null; then
        systemctl disable "$SERVICE_FILE"
        print_success "Service disabled"
    else
        print_warning "Service was not enabled"
    fi
}

# Function to remove systemd service file
remove_service_file() {
    print_status "Removing systemd service file..."

    if [[ -f "/etc/systemd/system/$SERVICE_FILE" ]]; then
        rm -f "/etc/systemd/system/$SERVICE_FILE"
        systemctl daemon-reload
        print_success "Service file removed and systemd reloaded"
    else
        print_warning "Service file not found"
    fi
}

# Function to remove installation directory
remove_install_dir() {
    print_status "Removing installation directory..."

    if [[ -d "$INSTALL_DIR" ]]; then
        # Backup images if they exist
        if [[ -d "$INSTALL_DIR/web/static/assets" ]] && [[ $(ls -A "$INSTALL_DIR/web/static/assets" 2>/dev/null | wc -l) -gt 0 ]]; then
            BACKUP_DIR="/tmp/kriti-images-backup-$(date +%Y%m%d-%H%M%S)"
            mkdir -p "$BACKUP_DIR"
            cp -r "$INSTALL_DIR/web/static/assets" "$BACKUP_DIR/"
            print_warning "Images backed up to: $BACKUP_DIR"
            print_warning "Please manually remove this backup when no longer needed"
        fi

        rm -rf "$INSTALL_DIR"
        print_success "Installation directory removed"
    else
        print_warning "Installation directory not found"
    fi
}

# Function to remove configuration directory
remove_config_dir() {
    print_status "Removing configuration directory..."

    if [[ -d "$CONFIG_DIR" ]]; then
        rm -rf "$CONFIG_DIR"
        print_success "Configuration directory removed"
    else
        print_warning "Configuration directory not found"
    fi
}

# Function to remove log directory
remove_log_dir() {
    print_status "Removing log directory..."

    if [[ -d "$LOG_DIR" ]]; then
        rm -rf "$LOG_DIR"
        print_success "Log directory removed"
    else
        print_warning "Log directory not found"
    fi
}

# Function to remove service user
remove_user() {
    print_status "Removing service user..."

    if id "$SERVICE_USER" &>/dev/null; then
        userdel "$SERVICE_USER" 2>/dev/null || true
        print_success "Service user removed"
    else
        print_warning "Service user not found"
    fi
}

# Function to remove logrotate configuration
remove_logrotate() {
    print_status "Removing logrotate configuration..."

    if [[ -f "/etc/logrotate.d/kriti-images" ]]; then
        rm -f "/etc/logrotate.d/kriti-images"
        print_success "Logrotate configuration removed"
    else
        print_warning "Logrotate configuration not found"
    fi
}

# Function to remove nginx configuration (if exists)
remove_nginx_config() {
    print_status "Checking for nginx configuration..."

    if [[ -f "/etc/nginx/sites-available/kriti-images" ]]; then
        rm -f "/etc/nginx/sites-available/kriti-images"
        if [[ -L "/etc/nginx/sites-enabled/kriti-images" ]]; then
            rm -f "/etc/nginx/sites-enabled/kriti-images"
        fi
        print_success "Nginx configuration removed"
        print_warning "Please reload nginx: systemctl reload nginx"
    else
        print_status "No nginx configuration found"
    fi
}

# Function to remove backup scripts
remove_backup_scripts() {
    print_status "Removing backup scripts..."

    if [[ -f "/usr/local/bin/backup-kriti-images.sh" ]]; then
        rm -f "/usr/local/bin/backup-kriti-images.sh"
        print_success "Backup script removed"
    fi

    # Remove from crontab
    if crontab -l 2>/dev/null | grep -q "backup-kriti-images.sh"; then
        crontab -l 2>/dev/null | grep -v "backup-kriti-images.sh" | crontab -
        print_success "Backup cron job removed"
    fi
}

# Function to clean up remaining processes
cleanup_processes() {
    print_status "Checking for remaining processes..."

    if pgrep -f "kriti-images" > /dev/null; then
        print_warning "Found running Kriti Images processes, attempting to stop them..."
        pkill -f "kriti-images" || true
        sleep 2

        if pgrep -f "kriti-images" > /dev/null; then
            print_warning "Force killing remaining processes..."
            pkill -9 -f "kriti-images" || true
        fi
        print_success "Processes cleaned up"
    else
        print_status "No running processes found"
    fi
}

# Function to show completion message
show_completion() {
    echo
    print_success "Kriti Images has been completely uninstalled!"
    echo
    print_status "What was removed:"
    echo "  ✓ Systemd service and configuration"
    echo "  ✓ Installation directory ($INSTALL_DIR)"
    echo "  ✓ Configuration directory ($CONFIG_DIR)"
    echo "  ✓ Log directory ($LOG_DIR)"
    echo "  ✓ Service user ($SERVICE_USER)"
    echo "  ✓ Logrotate configuration"
    echo "  ✓ Backup scripts and cron jobs"

    if [[ -d "/tmp/kriti-images-backup-"* ]]; then
        echo
        print_warning "Image backups are available in /tmp/"
        echo "Please manually remove them when no longer needed:"
        ls -la /tmp/kriti-images-backup-* 2>/dev/null || true
    fi

    echo
    print_status "You may also want to:"
    echo "  - Remove any nginx/apache configurations manually"
    echo "  - Close firewall ports (8080) if no longer needed"
    echo "  - Remove SSL certificates if they were specific to this service"
}

# Main uninstallation function
main() {
    echo -e "${RED}Kriti Images - Systemd Uninstallation Script${NC}"
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
            --force)
                FORCE_UNINSTALL="true"
                shift
                ;;
            --keep-data)
                KEEP_DATA="true"
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --install-dir DIR    Installation directory (default: /opt/kriti-images)"
                echo "  --user USER          Service user (default: kriti-images)"
                echo "  --force              Skip confirmation prompt"
                echo "  --keep-data          Keep user data and images (not implemented yet)"
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

    print_status "Uninstallation settings:"
    echo "  Install directory: $INSTALL_DIR"
    echo "  Service user: $SERVICE_USER"
    echo

    # Run uninstallation steps
    check_root
    confirm_uninstall
    stop_service
    cleanup_processes
    remove_service_file
    remove_install_dir
    remove_config_dir
    remove_log_dir
    remove_user
    remove_logrotate
    remove_nginx_config
    remove_backup_scripts
    show_completion
}

# Handle script interruption
trap 'echo -e "\n${RED}[ERROR]${NC} Uninstallation interrupted!"; exit 1' INT TERM

# Run main function
main "$@"
