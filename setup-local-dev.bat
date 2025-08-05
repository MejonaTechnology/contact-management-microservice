@echo off
echo ====================================================
echo Mejona Contact Microservice - Local Development Setup
echo ====================================================

echo.
echo 1. Checking Go installation...
go version
if %errorlevel% neq 0 (
    echo ERROR: Go is not installed or not in PATH
    echo Please install Go 1.23+ and try again
    pause
    exit /b 1
)

echo.
echo 2. Building Go microservice...
go build -o contact-service.exe cmd/server/main.go
if %errorlevel% neq 0 (
    echo ERROR: Failed to build Go microservice
    echo Check the error messages above
    pause
    exit /b 1
)

echo.
echo ✅ Build successful!
echo.
echo 3. Database Setup Information:
echo Please ensure MySQL is running and create the database:
echo   mysql -u root -p ^< test-db-setup.sql
echo.
echo 4. Environment Configuration:
echo Using .env.local for development settings
echo Edit .env.local to adjust database credentials if needed
echo.

echo 5. Starting contact microservice on port 8081...
echo Press Ctrl+C to stop the service
echo.
echo 🌐 Service will be available at:
echo   📊 Health Check: http://localhost:8081/health
echo   📧 Contact API: http://localhost:8081/api/v1/public/contact
echo   🧪 Test Endpoint: http://localhost:8081/api/v1/test
echo   📋 Admin API: http://localhost:8081/api/v1/dashboard/contacts
echo   📚 Swagger Docs: http://localhost:8081/swagger/index.html
echo.
echo 🧪 To test the API after starting:
echo   python test-contact-api.py
echo.

rem Use local development environment
set ENV_FILE=.env.local
if exist .env.local (
    echo Using .env.local for development
) else (
    echo Using .env for configuration
)

contact-service.exe