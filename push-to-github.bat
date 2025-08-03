@echo off
echo Pushing Contact Management Microservice to GitHub...
cd /d "D:\Mejona Workspace\Product\Mejona Admin Dashboard\services\contact-service"

echo Adding all files to git...
git add .

echo Committing final deployment-ready changes...
git commit -m "Final deployment preparation - Contact Management Microservice

✅ Complete Go microservice with 20 API endpoints
✅ Production-ready configuration
✅ AWS EC2 deployment automation
✅ Health monitoring and metrics
✅ JWT authentication system
✅ MySQL database integration
✅ Docker containerization
✅ Comprehensive documentation

🚀 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"

echo Pushing to GitHub...
git push -u origin main

echo ✅ Code successfully pushed to GitHub!
echo Repository URL: https://github.com/mejonatechnology/contact-management-microservice
pause