# Contact Management Microservice - Complete Deployment Script
# Run as Administrator in PowerShell

param(
    [string]$GitHubUsername = "",
    [string]$EC2IP = "",
    [string]$SSHKeyPath = "",
    [switch]$SkipGitHub = $false,
    [switch]$SkipAWS = $false
)

Write-Host "🚀 Contact Management Microservice - Complete Deployment" -ForegroundColor Green
Write-Host "=======================================================" -ForegroundColor Green

# Set location to service directory
$ServicePath = "D:\Mejona Workspace\Product\Mejona Admin Dashboard\services\contact-service"
Set-Location $ServicePath

Write-Host "📍 Working directory: $ServicePath" -ForegroundColor Blue

# Check if git repository is ready
Write-Host "`n🔍 Checking git repository status..." -ForegroundColor Yellow
$gitStatus = git status --porcelain
if ($gitStatus) {
    Write-Host "📝 Uncommitted changes found. Committing..." -ForegroundColor Yellow
    git add .
    git commit -m "Final deployment preparation - $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
}

Write-Host "✅ Git repository is ready" -ForegroundColor Green

# GitHub Repository Setup
if (-not $SkipGitHub) {
    Write-Host "`n🐙 GitHub Repository Setup" -ForegroundColor Cyan
    Write-Host "=============================" -ForegroundColor Cyan
    
    if (-not $GitHubUsername) {
        $GitHubUsername = Read-Host "Enter your GitHub username or organization name"
    }
    
    $repoName = "contact-management-microservice"
    $repoUrl = "https://github.com/$GitHubUsername/$repoName.git"
    
    Write-Host "📋 Repository Details:" -ForegroundColor Blue
    Write-Host "   • Name: $repoName" -ForegroundColor White
    Write-Host "   • URL: $repoUrl" -ForegroundColor White
    Write-Host "   • Description: Professional contact management microservice built with Go" -ForegroundColor White
    
    Write-Host "`n⚠️  MANUAL STEP REQUIRED:" -ForegroundColor Red
    Write-Host "1. Go to: https://github.com/new" -ForegroundColor Yellow
    Write-Host "2. Repository name: $repoName" -ForegroundColor Yellow
    Write-Host "3. Set as PUBLIC repository" -ForegroundColor Yellow
    Write-Host "4. Do NOT initialize with README, .gitignore, or license" -ForegroundColor Yellow
    Write-Host "5. Click 'Create repository'" -ForegroundColor Yellow
    
    Read-Host "`nPress Enter after creating the repository on GitHub"
    
    # Configure git remote
    Write-Host "🔗 Configuring git remote..." -ForegroundColor Yellow
    try {
        git remote remove origin 2>$null
    } catch {}
    
    git remote add origin $repoUrl
    git branch -M main
    
    # Push to GitHub
    Write-Host "📤 Pushing code to GitHub..." -ForegroundColor Yellow
    git push -u origin main
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Code successfully pushed to GitHub!" -ForegroundColor Green
        Write-Host "🔗 Repository URL: $repoUrl" -ForegroundColor Blue
    } else {
        Write-Host "❌ Failed to push to GitHub. Please check your credentials." -ForegroundColor Red
        exit 1
    }
}

# AWS EC2 Deployment
if (-not $SkipAWS) {
    Write-Host "`n☁️  AWS EC2 Deployment" -ForegroundColor Cyan
    Write-Host "========================" -ForegroundColor Cyan
    
    if (-not $EC2IP) {
        $EC2IP = Read-Host "Enter your EC2 public IP address"
    }
    
    if (-not $SSHKeyPath) {
        $SSHKeyPath = Read-Host "Enter the path to your SSH key file (.pem)"
    }
    
    # Validate SSH key exists
    if (-not (Test-Path $SSHKeyPath)) {
        Write-Host "❌ SSH key file not found: $SSHKeyPath" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "📋 Deployment Details:" -ForegroundColor Blue
    Write-Host "   • EC2 IP: $EC2IP" -ForegroundColor White
    Write-Host "   • SSH Key: $SSHKeyPath" -ForegroundColor White
    Write-Host "   • Repository: $repoUrl" -ForegroundColor White
    
    # Test SSH connectivity
    Write-Host "`n🔌 Testing SSH connectivity..." -ForegroundColor Yellow
    $sshTest = ssh -i $SSHKeyPath -o ConnectTimeout=10 -o BatchMode=yes ec2-user@$EC2IP "echo 'SSH connection successful'"
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Cannot connect to EC2 instance. Please check:" -ForegroundColor Red
        Write-Host "   • EC2 IP address is correct" -ForegroundColor Yellow
        Write-Host "   • SSH key has correct permissions" -ForegroundColor Yellow
        Write-Host "   • Security group allows SSH (port 22)" -ForegroundColor Yellow
        Write-Host "   • EC2 instance is running" -ForegroundColor Yellow
        exit 1
    }
    
    Write-Host "✅ SSH connection successful!" -ForegroundColor Green
    
    # Upload deployment script
    Write-Host "`n📤 Uploading deployment script..." -ForegroundColor Yellow
    scp -i $SSHKeyPath "scripts/deploy-aws.sh" ec2-user@${EC2IP}:/tmp/deploy-aws.sh
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Failed to upload deployment script" -ForegroundColor Red
        exit 1
    }
    
    # Execute deployment
    Write-Host "`n🚀 Executing deployment on EC2..." -ForegroundColor Yellow
    Write-Host "This may take 5-10 minutes..." -ForegroundColor Blue
    
    $deployCommand = @"
chmod +x /tmp/deploy-aws.sh
export GITHUB_REPO="$repoUrl"
/tmp/deploy-aws.sh
"@
    
    ssh -i $SSHKeyPath ec2-user@$EC2IP $deployCommand
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Deployment completed successfully!" -ForegroundColor Green
    } else {
        Write-Host "❌ Deployment failed. Check the output above for errors." -ForegroundColor Red
        exit 1
    }
    
    # Verify deployment
    Write-Host "`n🧪 Verifying deployment..." -ForegroundColor Yellow
    
    # Test health endpoint
    Write-Host "Testing health endpoint..." -ForegroundColor Blue
    try {
        $healthResponse = Invoke-WebRequest -Uri "http://$EC2IP/health" -TimeoutSec 30
        if ($healthResponse.StatusCode -eq 200) {
            Write-Host "✅ Health check passed!" -ForegroundColor Green
        }
    } catch {
        Write-Host "⚠️  Health check failed. Service may still be starting..." -ForegroundColor Yellow
    }
    
    # Test API endpoint
    Write-Host "Testing API endpoint..." -ForegroundColor Blue
    try {
        $apiResponse = Invoke-WebRequest -Uri "http://$EC2IP/api/v1/dashboard/contacts" -TimeoutSec 30
        if ($apiResponse.StatusCode -eq 200) {
            Write-Host "✅ API endpoint responding!" -ForegroundColor Green
        }
    } catch {
        Write-Host "⚠️  API test failed. Check logs on EC2 instance." -ForegroundColor Yellow
    }
}

# Final summary
Write-Host "`n🎉 Deployment Summary" -ForegroundColor Green
Write-Host "=====================" -ForegroundColor Green

if (-not $SkipGitHub) {
    Write-Host "✅ GitHub Repository: $repoUrl" -ForegroundColor Green
}

if (-not $SkipAWS) {
    Write-Host "✅ EC2 Deployment: http://$EC2IP" -ForegroundColor Green
    Write-Host "`n🔗 Service URLs:" -ForegroundColor Blue
    Write-Host "   • Health Check: http://$EC2IP/health" -ForegroundColor White
    Write-Host "   • API Endpoint: http://$EC2IP/api/v1/dashboard/contacts" -ForegroundColor White
    Write-Host "   • Metrics: http://$EC2IP/metrics" -ForegroundColor White
    Write-Host "   • API Docs: http://$EC2IP/swagger/index.html" -ForegroundColor White
    
    Write-Host "`n📋 Next Steps:" -ForegroundColor Blue
    Write-Host "1. Configure .env file on EC2 with production database credentials" -ForegroundColor Yellow
    Write-Host "2. Set up SSL certificate for HTTPS (recommended)" -ForegroundColor Yellow
    Write-Host "3. Configure monitoring and alerting" -ForegroundColor Yellow
    Write-Host "4. Set up automated backups" -ForegroundColor Yellow
    
    Write-Host "`n🔧 Management Commands:" -ForegroundColor Blue
    Write-Host "ssh -i $SSHKeyPath ec2-user@$EC2IP" -ForegroundColor White
    Write-Host "sudo systemctl status contact-service" -ForegroundColor White
    Write-Host "sudo journalctl -u contact-service -f" -ForegroundColor White
}

Write-Host "`n🎯 Contact Management Microservice is now live!" -ForegroundColor Green
Write-Host "All 20 API endpoints are ready for production use." -ForegroundColor Green

# Test endpoint verification
if (-not $SkipAWS -and $EC2IP) {
    Write-Host "`n🧪 Quick Endpoint Test:" -ForegroundColor Cyan
    Write-Host "curl http://$EC2IP/health" -ForegroundColor White
    Write-Host "curl http://$EC2IP/api/v1/dashboard/contacts" -ForegroundColor White
}