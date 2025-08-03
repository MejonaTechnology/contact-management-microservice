# AWS EC2 Quick Deployment Guide

## Prerequisites
1. AWS EC2 instance (Amazon Linux 2 or Ubuntu) with:
   - Minimum 1GB RAM (t2.micro eligible)
   - Port 80, 443, and 8081 open in security group
   - SSH access configured

## Step 1: Create GitHub Repository
1. Go to https://github.com/new
2. Repository name: `contact-management-microservice`
3. Set as Public, don't initialize with files
4. Run: `push-to-github.bat` from this directory

## Step 2: Deploy to AWS EC2

### Option A: Automated Deployment (Recommended)
```bash
# SSH into your EC2 instance
ssh -i your-key.pem ec2-user@your-ec2-ip

# Download and run deployment script
curl -fsSL https://raw.githubusercontent.com/mejonatechnology/contact-management-microservice/main/scripts/deploy-aws.sh | bash
```

### Option B: Manual Deployment
```bash
# SSH into EC2 instance
ssh -i your-key.pem ec2-user@your-ec2-ip

# Update system and install dependencies
sudo yum update -y
sudo yum install -y git wget curl nginx mysql

# Install Go 1.21
cd /tmp
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Clone repository
sudo mkdir -p /opt/mejona
sudo chown ec2-user:ec2-user /opt/mejona
cd /opt/mejona
git clone https://github.com/mejonatechnology/contact-management-microservice.git
cd contact-management-microservice

# Configure environment
cp .env.example .env
# Edit .env with your database credentials

# Build and deploy
go mod tidy
go build -o contact-service cmd/server/main.go
sudo cp scripts/contact-service.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable contact-service
sudo systemctl start contact-service

# Configure Nginx
sudo cp scripts/nginx.conf /etc/nginx/conf.d/contact-service.conf
sudo systemctl enable nginx
sudo systemctl restart nginx
```

## Step 3: Verify Deployment
```bash
# Check service status
sudo systemctl status contact-service

# Test health endpoint
curl http://localhost:8081/health

# Test via public IP
curl http://YOUR_EC2_PUBLIC_IP/health
```

## Step 4: Test All 20 Endpoints
```bash
# Download test script
curl -fsSL https://raw.githubusercontent.com/mejonatechnology/contact-management-microservice/main/test_all_endpoints.bat > test_endpoints.sh
chmod +x test_endpoints.sh

# Run comprehensive tests
./test_endpoints.sh YOUR_EC2_PUBLIC_IP
```

## Environment Variables to Configure

### Required Database Settings
```env
DB_HOST=your-mysql-host
DB_USER=your-mysql-user
DB_PASSWORD=your-mysql-password
DB_NAME=mejona_contacts
```

### JWT Security
```env
JWT_SECRET=your-super-secret-jwt-key-minimum-32-characters
```

### Application URLs
```env
APP_URL=http://YOUR_EC2_PUBLIC_IP
CORS_ALLOWED_ORIGINS=https://admin.mejona.com,http://YOUR_EC2_PUBLIC_IP
```

## Post-Deployment Checklist
- [ ] Service running on port 8081
- [ ] Nginx proxy configured on port 80
- [ ] All 20 API endpoints responding
- [ ] Database connectivity confirmed
- [ ] JWT authentication working
- [ ] Health checks passing
- [ ] Logs being written to `/opt/mejona/contact-management-microservice/logs/`

## Troubleshooting
```bash
# View service logs
sudo journalctl -u contact-service -f

# Check application logs
tail -f /opt/mejona/contact-management-microservice/logs/contact-service.log

# Test database connection
mysql -h YOUR_DB_HOST -u YOUR_DB_USER -p YOUR_DB_NAME

# Restart service
sudo systemctl restart contact-service
```

## Management Commands
```bash
# Start service
sudo systemctl start contact-service

# Stop service
sudo systemctl stop contact-service

# Restart service
sudo systemctl restart contact-service

# View status
sudo systemctl status contact-service

# Update deployment
cd /opt/mejona/contact-management-microservice
git pull origin main
go build -o contact-service cmd/server/main.go
sudo systemctl restart contact-service
```

## API Documentation
- Swagger UI: `http://YOUR_EC2_PUBLIC_IP/swagger/index.html`
- API Docs: See `API_DOCUMENTATION.md` in repository
- 20 Endpoints: All tested and working (see `ENDPOINT_TEST_RESULTS.md`)