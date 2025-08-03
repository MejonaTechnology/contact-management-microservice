@echo off
echo Starting Mejona Contact Service...
echo.

REM Check if Go is installed
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Go is not installed or not in PATH
    pause
    exit /b 1
)

REM Build and run the service
echo Building contact service...
go build -o contact-service.exe cmd/server/main.go
if %errorlevel% neq 0 (
    echo Error: Failed to build contact service
    pause
    exit /b 1
)

echo Starting contact service on port 8081...
echo Dashboard endpoints will be available at:
echo   GET  http://localhost:8081/api/v1/dashboard/contacts
echo   GET  http://localhost:8081/api/v1/dashboard/contacts/stats
echo   POST http://localhost:8081/api/v1/dashboard/contact
echo   PUT  http://localhost:8081/api/v1/dashboard/contacts/{id}/status
echo.
echo Press Ctrl+C to stop the service
echo.

contact-service.exe