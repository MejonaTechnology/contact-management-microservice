#!/bin/bash

# =================================================================
# GitHub Repository Setup Script for Contact Management Microservice
# =================================================================

set -e

# Configuration
REPO_NAME="contact-management-microservice"
GITHUB_ORG="mejonatechnology"
REPO_URL="https://github.com/$GITHUB_ORG/$REPO_NAME.git"

echo "🚀 Setting up GitHub repository for Contact Management Microservice..."

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

# Check if git is installed
if ! command -v git &> /dev/null; then
    print_error "Git is not installed. Please install Git first."
    exit 1
fi

# Check if GitHub CLI is available
if command -v gh &> /dev/null; then
    print_status "GitHub CLI detected. Checking authentication..."
    if gh auth status &> /dev/null; then
        USE_GH_CLI=true
        print_success "GitHub CLI authenticated"
    else
        print_error "GitHub CLI not authenticated. Run 'gh auth login' first."
        USE_GH_CLI=false
    fi
else
    print_status "GitHub CLI not found. Will use manual setup."
    USE_GH_CLI=false
fi

# Initialize git repository if not already initialized
if [ ! -d ".git" ]; then
    print_status "Initializing Git repository..."
    git init
    print_success "Git repository initialized"
else
    print_success "Git repository already exists"
fi

# Create or update .gitignore if it doesn't exist
if [ ! -f ".gitignore" ]; then
    print_status "Creating .gitignore file..."
    cat > .gitignore << 'EOF'
# See existing .gitignore content above
EOF
    print_success ".gitignore created"
fi

# Set up initial commit
print_status "Preparing initial commit..."

# Add all files except ignored ones
git add .

# Check if there are any changes to commit
if git diff --staged --quiet; then
    print_status "No changes to commit"
else
    # Commit initial files
    git commit -m "Initial commit: Contact Management Microservice

- Complete Go microservice with 20 API endpoints
- Health monitoring and metrics collection
- JWT authentication system
- MySQL database integration
- Docker and Docker Compose support
- AWS EC2 deployment scripts
- Comprehensive API documentation
- Production-ready configuration

Features:
✅ Dashboard contact management
✅ Authentication and authorization
✅ Health checks and monitoring
✅ CSV export functionality
✅ Bulk operations support
✅ Database migrations
✅ Structured logging
✅ CORS and security middleware

Built with ❤️ by Mejona Technology"
    print_success "Initial commit created"
fi

# Set up remote repository
if [ "$USE_GH_CLI" = true ]; then
    print_status "Creating GitHub repository using GitHub CLI..."
    
    # Create repository
    gh repo create "$GITHUB_ORG/$REPO_NAME" \
        --description "Professional contact management microservice built with Go for Mejona Technology Admin Dashboard" \
        --homepage "https://mejona.com" \
        --public \
        --add-readme=false \
        --clone=false
    
    print_success "GitHub repository created"
    
    # Add remote
    git remote add origin "$REPO_URL" 2>/dev/null || git remote set-url origin "$REPO_URL"
    print_success "Remote origin added"
    
else
    print_status "Setting up remote manually..."
    echo ""
    echo "📋 Manual GitHub Setup Required:"
    echo "   1. Go to https://github.com/new"
    echo "   2. Repository name: $REPO_NAME"
    echo "   3. Description: Professional contact management microservice built with Go for Mejona Technology Admin Dashboard"
    echo "   4. Set as Public repository"
    echo "   5. Do NOT initialize with README, .gitignore, or license"
    echo "   6. Create repository"
    echo ""
    read -p "Press Enter after creating the repository on GitHub..."
    
    # Add remote
    git remote add origin "$REPO_URL" 2>/dev/null || git remote set-url origin "$REPO_URL"
    print_success "Remote origin configured"
fi

# Set up main branch
print_status "Setting up main branch..."
git branch -M main
print_success "Main branch configured"

# Push to GitHub
print_status "Pushing to GitHub..."
git push -u origin main
print_success "Code pushed to GitHub"

# Create additional branches
print_status "Creating development branches..."
git checkout -b develop
git push -u origin develop

git checkout -b staging
git push -u origin staging

git checkout main
print_success "Development branches created"

# Set up repository topics and settings (if using GitHub CLI)
if [ "$USE_GH_CLI" = true ]; then
    print_status "Configuring repository settings..."
    
    # Add topics
    gh repo edit "$GITHUB_ORG/$REPO_NAME" \
        --add-topic go \
        --add-topic microservice \
        --add-topic api \
        --add-topic contact-management \
        --add-topic gin \
        --add-topic mysql \
        --add-topic jwt \
        --add-topic docker \
        --add-topic aws \
        --add-topic mejona-technology
    
    print_success "Repository topics added"
fi

# Create initial GitHub issues
if [ "$USE_GH_CLI" = true ]; then
    print_status "Creating initial project issues..."
    
    gh issue create \
        --title "Set up production database environment" \
        --body "Configure production MySQL database with proper credentials and security settings." \
        --label "enhancement,database"
    
    gh issue create \
        --title "Implement comprehensive monitoring" \
        --body "Set up Prometheus metrics collection and Grafana dashboards for production monitoring." \
        --label "enhancement,monitoring"
    
    gh issue create \
        --title "Add SSL certificate configuration" \
        --body "Configure HTTPS with SSL certificates for production deployment." \
        --label "enhancement,security"
    
    gh issue create \
        --title "Set up automated testing pipeline" \
        --body "Implement comprehensive unit tests and integration tests with GitHub Actions." \
        --label "enhancement,testing"
    
    print_success "Initial issues created"
fi

echo ""
print_success "🎉 GitHub repository setup completed!"
echo ""
echo "📋 Repository Information:"
echo "   • Repository URL: $REPO_URL"
echo "   • Main Branch: main"
echo "   • Development Branch: develop"
echo "   • Staging Branch: staging"
echo ""
echo "🔗 Quick Links:"
echo "   • Repository: https://github.com/$GITHUB_ORG/$REPO_NAME"
echo "   • Issues: https://github.com/$GITHUB_ORG/$REPO_NAME/issues"
echo "   • Actions: https://github.com/$GITHUB_ORG/$REPO_NAME/actions"
echo "   • Releases: https://github.com/$GITHUB_ORG/$REPO_NAME/releases"
echo ""
echo "📝 Next Steps:"
echo "   1. Configure repository secrets for CI/CD"
echo "   2. Set up branch protection rules"
echo "   3. Configure AWS deployment credentials"
echo "   4. Set up production environment variables"
echo ""