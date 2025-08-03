# Contact Management Microservice - Deployment Guide

## 🚀 Complete Deployment Instructions

### Overview
This guide provides step-by-step instructions for deploying the Contact Management Microservice to production environments, including GitHub repository setup and AWS EC2 deployment.

### Prerequisites
- AWS EC2 instance (Amazon Linux 2 or Ubuntu)
- GitHub account with repository access
- Database server (MySQL 8.0+)
- Domain name (optional, for SSL)

## 📊 Service Status
- **Total API Endpoints**: 20
- **Working Endpoints**: 20/20 ✅
- **Test Coverage**: Comprehensive
- **Production Ready**: ✅

## 🔧 GitHub Repository Setup

### Step 1: Create GitHub Repository
1. Go to [GitHub New Repository](https://github.com/new)
2. **Repository name**: `contact-management-microservice`
3. **Description**: `Professional contact management microservice built with Go for Mejona Technology Admin Dashboard`
4. **Visibility**: Public
5. **Initialize**: Do NOT check any initialization options
6. Click **Create repository**

### Step 2: Push Local Code
```bash
# Navigate to project directory
cd "D:\Mejona Workspace\Product\Mejona Admin Dashboard\services\contact-service"

# Add remote origin (replace with your GitHub username/org)
git remote add origin https://github.com/YOUR_USERNAME/contact-management-microservice.git

# Rename branch to main
git branch -M main

# Push to GitHub
git push -u origin main

# Create development branches
git checkout -b develop
git push -u origin develop

git checkout -b staging  
git push -u origin staging

git checkout main
```

## ☁️ AWS EC2 Deployment

### Step 1: Launch EC2 Instance
1. **Instance Type**: t3.micro or larger
2. **OS**: Amazon Linux 2 
3. **Security Groups**: 
   - SSH (22) - Your IP
   - HTTP (80) - Anywhere
   - HTTPS (443) - Anywhere
   - Custom (8081) - Anywhere (for direct API access)
4. **Storage**: 20GB GP3
5. **Key Pair**: Create/select for SSH access

### Step 2: Connect to EC2 Instance
```bash
# Connect via SSH
ssh -i your-key.pem ec2-user@YOUR_EC2_PUBLIC_IP
```

### Step 3: Run Deployment Script
```bash
# Download and execute deployment script
curl -O https://raw.githubusercontent.com/YOUR_USERNAME/contact-management-microservice/main/scripts/deploy-aws.sh
chmod +x deploy-aws.sh
./deploy-aws.sh
```

### Step 4: Configure Environment
```bash
# Navigate to deployment directory
cd /opt/mejona/contact-management-microservice

# Edit environment file
sudo nano .env

# Update with your configuration:
DB_HOST=your-database-host
DB_PORT=3306
DB_USER=your-db-user
DB_PASSWORD=your-db-password
DB_NAME=contact_management
JWT_SECRET=your-super-secure-jwt-secret-key
```

### Step 5: Start Services
```bash
# Start the contact service
sudo systemctl start contact-service
sudo systemctl enable contact-service

# Check service status
sudo systemctl status contact-service

# Start Nginx
sudo systemctl start nginx
sudo systemctl enable nginx
```

## 🧪 Verification & Testing

### Health Check
```bash
# Local health check
curl http://localhost:8081/health

# Public health check
curl http://YOUR_EC2_PUBLIC_IP/health
```

### API Endpoint Testing
```bash
# Test contact creation
curl -X POST http://YOUR_EC2_PUBLIC_IP/api/v1/dashboard/contact \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "message": "Test message"
  }'

# Test contact retrieval
curl http://YOUR_EC2_PUBLIC_IP/api/v1/dashboard/contacts
```

## 📋 All Available Endpoints

### Dashboard Endpoints (5)
1. `GET /api/v1/dashboard/contacts` - Get all contact submissions
2. `POST /api/v1/dashboard/contact` - Create new contact submission
3. `GET /api/v1/dashboard/contacts/:id` - Get specific contact
4. `PUT /api/v1/dashboard/contacts/:id/status` - Update contact status
5. `GET /api/v1/dashboard/contacts/stats` - Get contact statistics

### Contact Management Endpoints (10)
6. `GET /api/v1/contacts` - List all contacts
7. `POST /api/v1/contacts` - Create new contact
8. `GET /api/v1/contacts/:id` - Get contact by ID
9. `PUT /api/v1/contacts/:id` - Update contact
10. `DELETE /api/v1/contacts/:id` - Delete contact
11. `GET /api/v1/contacts/search` - Search contacts
12. `GET /api/v1/contacts/export` - Export contacts to CSV
13. `POST /api/v1/contacts/bulk` - Bulk operations
14. `GET /api/v1/contacts/:id/history` - Contact history
15. `PUT /api/v1/contacts/:id/status` - Update contact status

### System Endpoints (5)
16. `GET /health` - Health check
17. `GET /metrics` - Prometheus metrics
18. `GET /api/v1/health/detailed` - Detailed health info
19. `POST /api/v1/auth/login` - Authentication
20. `POST /api/v1/auth/refresh` - Token refresh

## 🔐 Security Configuration

### SSL Certificate Setup (Optional)
```bash
# Install Certbot
sudo yum install -y certbot python3-certbot-nginx

# Get SSL certificate
sudo certbot --nginx -d your-domain.com

# Auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

### Firewall Configuration
```bash
# Configure AWS Security Groups to allow:
# - Port 22 (SSH) from your IP only
# - Port 80 (HTTP) from anywhere
# - Port 443 (HTTPS) from anywhere
# - Port 8081 (API) from anywhere (optional, can route through Nginx)
```

## 📊 Monitoring & Maintenance

### Service Management
```bash
# View service logs
sudo journalctl -u contact-service -f

# Restart service
sudo systemctl restart contact-service

# Check service status
sudo systemctl status contact-service
```

### Performance Monitoring
```bash
# System resources
htop

# Service metrics
curl http://localhost:8081/metrics

# Database connections
mysql -h YOUR_DB_HOST -u YOUR_DB_USER -p -e "SHOW PROCESSLIST;"
```

### Log Management
```bash
# View application logs
tail -f /opt/mejona/contact-management-microservice/logs/app.log

# View Nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

## 🔄 Updates & Maintenance

### Deploy Updates
```bash
cd /opt/mejona/contact-management-microservice
./deploy-update.sh
```

### Database Migrations
```bash
# Run pending migrations
cd /opt/mejona/contact-management-microservice
go run migrations/migrate.go
```

### Backup Strategy
```bash
# Database backup
mysqldump -h YOUR_DB_HOST -u YOUR_DB_USER -p contact_management > backup_$(date +%Y%m%d).sql

# Application backup
tar -czf contact-service-backup-$(date +%Y%m%d).tar.gz /opt/mejona/contact-management-microservice
```

## 🚨 Troubleshooting

### Common Issues

1. **Service won't start**
   ```bash
   sudo journalctl -u contact-service -n 50
   ```

2. **Database connection failed**
   ```bash
   # Test database connectivity
   mysql -h YOUR_DB_HOST -u YOUR_DB_USER -p
   ```

3. **Port already in use**
   ```bash
   sudo netstat -tlnp | grep :8081
   sudo kill -9 PID_NUMBER
   ```

4. **Permission denied**
   ```bash
   sudo chown -R ec2-user:ec2-user /opt/mejona/contact-management-microservice
   ```

### Support
- **GitHub Issues**: [Repository Issues](https://github.com/YOUR_USERNAME/contact-management-microservice/issues)
- **Documentation**: [API Documentation](./API_DOCUMENTATION.md)
- **Email**: support@mejona.com

## 📈 Production Checklist

- [ ] GitHub repository created and code pushed
- [ ] EC2 instance launched and configured
- [ ] Database credentials configured
- [ ] All 20 API endpoints tested
- [ ] SSL certificate installed (recommended)
- [ ] Monitoring and logging configured
- [ ] Backup strategy implemented
- [ ] Security groups configured
- [ ] Domain name configured (optional)
- [ ] Load balancer configured (for high availability)

---

**Built with ❤️ by Mejona Technology**

**Contact**: 
- Website: [mejona.com](https://mejona.com)
- Email: info@mejona.com
- Phone: +91 9546805580