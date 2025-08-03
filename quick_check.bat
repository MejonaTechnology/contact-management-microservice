@echo off
echo Testing AWS EC2 Connectivity and Service Status...
echo.

echo 1. Testing SSH connectivity (port 22)...
powershell -Command "if (Test-NetConnection -ComputerName 65.1.94.25 -Port 22 -InformationLevel Quiet) { Write-Host 'SSH Port 22: ACCESSIBLE' -ForegroundColor Green } else { Write-Host 'SSH Port 22: NOT ACCESSIBLE' -ForegroundColor Red }"

echo.
echo 2. Testing HTTP service (port 8081)...
powershell -Command "if (Test-NetConnection -ComputerName 65.1.94.25 -Port 8081 -InformationLevel Quiet) { Write-Host 'HTTP Port 8081: ACCESSIBLE' -ForegroundColor Green } else { Write-Host 'HTTP Port 8081: NOT ACCESSIBLE' -ForegroundColor Red }"

echo.
echo 3. Testing health endpoint...
powershell -Command "try { $response = Invoke-RestMethod -Uri 'http://65.1.94.25:8081/health' -TimeoutSec 5 -ErrorAction Stop; Write-Host 'Health Endpoint: RESPONDING' -ForegroundColor Green; Write-Host 'Service Status:' $response.data.status -ForegroundColor Cyan; Write-Host 'Service Name:' $response.data.service -ForegroundColor Cyan } catch { Write-Host 'Health Endpoint: NOT RESPONDING' -ForegroundColor Red; Write-Host 'Error:' $_.Exception.Message -ForegroundColor Yellow }"

echo.
echo 4. Testing SSH login and service status...
ssh -i "D:\Mejona Workspace\Mejona Cred\AWS\mejona.pem" -o ConnectTimeout=10 -o StrictHostKeyChecking=no ubuntu@65.1.94.25 "echo 'SSH Login: SUCCESS' && systemctl is-active contact-service && echo 'Service is active' || echo 'Service not running'"

echo.
echo Connectivity check completed.
pause