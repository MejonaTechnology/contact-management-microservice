#!/bin/bash

# =================================================================
# AWS EC2 Deployment Script for Contact Management Microservice
# =================================================================

set -e

# Configuration
PROJECT_NAME="contact-management-microservice"
GITHUB_REPO="https://github.com/mejonatechnology/contact-management-microservice.git"
SERVICE_PORT="8081"
EC2_USER="ec2-user"
DEPLOY_PATH="/opt/mejona"

echo "🚀 Starting AWS EC2 deployment for Contact Management Microservice..."

# Function to print colored output
print_status() {
    echo -e "\n\033[1;34m==>\033[0m $1"
}

print_success() {
    echo -e "\033[1;32m✓\033[0m $1"
}

print_error() {
    echo -e "\033[1;31m✗\033[0m $1"
}

# Check if running on EC2
if [ ! -f /sys/hypervisor/uuid ] || [ "$(head -c 3 /sys/hypervisor/uuid 2>/dev/null)" != "ec2" ]; then
    print_error "This script is designed to run on AWS EC2 instances"
    exit 1
fi

# Update system packages
print_status "Updating system packages..."
sudo yum update -y
print_success "System packages updated"

# Install required packages
print_status "Installing required packages..."
sudo yum install -y git wget curl nginx mysql

# Install Go
print_status "Installing Go 1.21..."
if ! command -v go &> /dev/null; then
    cd /tmp
    wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    source ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
    print_success "Go 1.21 installed"
else
    print_success "Go already installed"
fi

# Install Docker
print_status "Installing Docker..."
if ! command -v docker &> /dev/null; then
    sudo yum install -y docker
    sudo service docker start
    sudo usermod -a -G docker $EC2_USER
    sudo chkconfig docker on
    print_success "Docker installed and started"
else
    print_success "Docker already installed"
fi

# Install Docker Compose
print_status "Installing Docker Compose..."
if ! command -v docker-compose &> /dev/null; then
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
    print_success "Docker Compose installed"
else
    print_success "Docker Compose already installed"
fi

# Create deployment directory
print_status "Setting up deployment directory..."
sudo mkdir -p $DEPLOY_PATH
sudo chown $EC2_USER:$EC2_USER $DEPLOY_PATH
cd $DEPLOY_PATH

# Clone or update repository
print_status "Cloning/updating repository..."
if [ -d "$PROJECT_NAME" ]; then
    cd $PROJECT_NAME
    git pull origin main
    print_success "Repository updated"
else
    git clone $GITHUB_REPO $PROJECT_NAME
    cd $PROJECT_NAME
    print_success "Repository cloned"
fi

# Setup environment file
print_status "Setting up environment configuration..."
if [ ! -f .env ]; then
    cp .env.example .env
    echo "⚠️  Please update .env file with your configuration before starting the service"
    echo "   Edit: $DEPLOY_PATH/$PROJECT_NAME/.env"
fi

# Build the application
print_status "Building the application..."
go mod tidy
go build -o contact-service cmd/server/main.go
print_success "Application built successfully"

# Create systemd service
print_status "Creating systemd service..."
sudo tee /etc/systemd/system/contact-service.service > /dev/null <<EOF
[Unit]
Description=Mejona Contact Management Microservice
After=network.target

[Service]
Type=simple
User=$EC2_USER
WorkingDirectory=$DEPLOY_PATH/$PROJECT_NAME
ExecStart=$DEPLOY_PATH/$PROJECT_NAME/contact-service
Restart=always
RestartSec=5
Environment=PATH=/usr/local/go/bin:/usr/bin:/bin
EnvironmentFile=$DEPLOY_PATH/$PROJECT_NAME/.env

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=$DEPLOY_PATH/$PROJECT_NAME/logs

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and enable service
sudo systemctl daemon-reload
sudo systemctl enable contact-service
print_success "Systemd service created and enabled"

# Setup Nginx reverse proxy
print_status "Configuring Nginx reverse proxy..."
sudo tee /etc/nginx/conf.d/contact-service.conf > /dev/null <<EOF
upstream contact_service {
    server 127.0.0.1:$SERVICE_PORT;
}

server {
    listen 80;
    server_name _;
    
    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    
    # Rate limiting
    limit_req_zone \$binary_remote_addr zone=api:10m rate=10r/s;
    
    location / {
        limit_req zone=api burst=20 nodelay;
        
        proxy_pass http://contact_service;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # Health check endpoint
    location /health {
        proxy_pass http://contact_service/health;
        access_log off;
    }
}
EOF

# Test Nginx configuration
sudo nginx -t
if [ $? -eq 0 ]; then
    sudo systemctl enable nginx
    sudo systemctl restart nginx
    print_success "Nginx configured and started"
else
    print_error "Nginx configuration test failed"
    exit 1
fi

# Setup log rotation
print_status "Setting up log rotation..."
sudo tee /etc/logrotate.d/contact-service > /dev/null <<EOF
$DEPLOY_PATH/$PROJECT_NAME/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 $EC2_USER $EC2_USER
    postrotate
        systemctl reload contact-service > /dev/null 2>&1 || true
    endscript
}
EOF
print_success "Log rotation configured"

# Setup firewall rules (if firewalld is available)
if command -v firewall-cmd &> /dev/null; then
    print_status "Configuring firewall..."
    sudo firewall-cmd --permanent --add-service=http
    sudo firewall-cmd --permanent --add-service=https
    sudo firewall-cmd --permanent --add-port=$SERVICE_PORT/tcp
    sudo firewall-cmd --reload
    print_success "Firewall configured"
fi

# Create deployment script for updates
print_status "Creating update script..."
tee deploy-update.sh > /dev/null <<EOF
#!/bin/bash
cd $DEPLOY_PATH/$PROJECT_NAME
git pull origin main
go build -o contact-service cmd/server/main.go
sudo systemctl restart contact-service
echo "✓ Service updated and restarted"
EOF
chmod +x deploy-update.sh
print_success "Update script created"

# Final status check
print_status "Deployment completed! Performing final checks..."

# Start the service
sudo systemctl start contact-service

# Wait a moment for service to start
sleep 5

# Check service status
if sudo systemctl is-active --quiet contact-service; then
    print_success "Contact service is running"
else
    print_error "Contact service failed to start"
    echo "Check logs with: sudo journalctl -u contact-service -f"
    exit 1
fi

# Check if port is accessible
if curl -f -s http://localhost:$SERVICE_PORT/health > /dev/null; then
    print_success "Service health check passed"
else
    print_error "Service health check failed"
    exit 1
fi

# Get public IP
PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4 2>/dev/null || echo "N/A")

echo ""
echo "🎉 Deployment completed successfully!"
echo ""
echo "📋 Service Information:"
echo "   • Service Status: Running"
echo "   • Local URL: http://localhost:$SERVICE_PORT"
echo "   • Public URL: http://$PUBLIC_IP"
echo "   • Health Check: http://$PUBLIC_IP/health"
echo "   • API Documentation: http://$PUBLIC_IP/swagger/index.html"
echo ""
echo "📊 Management Commands:"
echo "   • Start Service: sudo systemctl start contact-service"
echo "   • Stop Service: sudo systemctl stop contact-service"
echo "   • Restart Service: sudo systemctl restart contact-service"
echo "   • View Logs: sudo journalctl -u contact-service -f"
echo "   • Update Service: ./deploy-update.sh"
echo ""
echo "⚠️  Next Steps:"
echo "   1. Update .env file with your database credentials"
echo "   2. Configure SSL certificate for HTTPS (recommended)"
echo "   3. Set up monitoring and alerting"
echo "   4. Configure database backups"
echo ""