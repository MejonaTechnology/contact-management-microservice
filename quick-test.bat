@echo off
echo =============================================
echo Quick Test - Contact Microservice
echo =============================================

echo.
echo 1. Testing if service is running...
curl -s http://localhost:8081/health > nul
if %errorlevel% neq 0 (
    echo ❌ Service is not running on port 8081
    echo Please start the service first: ./setup-local-dev.bat
    pause
    exit /b 1
)

echo ✅ Service is running!
echo.

echo 2. Testing health endpoint...
curl -s http://localhost:8081/health
echo.
echo.

echo 3. Testing contact submission...
curl -X POST http://localhost:8081/api/v1/public/contact ^
  -H "Content-Type: application/json" ^
  -d "{\"name\": \"Test User\", \"email\": \"test@example.com\", \"phone\": \"+1234567890\", \"message\": \"Test message from batch script\", \"subject\": \"Test\", \"source\": \"batch_test\", \"website\": \"\"}"

echo.
echo.
echo =============================================
echo Test completed!
echo.
echo 📊 Check the logs above for results
echo 🌐 Health endpoint: http://localhost:8081/health
echo 📧 Contact API: http://localhost:8081/api/v1/public/contact
echo.
pause