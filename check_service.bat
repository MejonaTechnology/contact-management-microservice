@echo off
echo Checking Contact Management Service Status on AWS EC2...
echo.

echo 1. SSH Connection Test...
ssh -i "D:\Mejona Workspace\Mejona Cred\AWS\mejona.pem" -o ConnectTimeout=5 -o StrictHostKeyChecking=no ubuntu@65.1.94.25 "echo 'SSH Connection: SUCCESS'"

echo.
echo 2. Service Status Check...
ssh -i "D:\Mejona Workspace\Mejona Cred\AWS\mejona.pem" -o ConnectTimeout=10 -o StrictHostKeyChecking=no ubuntu@65.1.94.25 "systemctl status contact-service --no-pager"

echo.
echo 3. Service Logs (Last 20 lines)...
ssh -i "D:\Mejona Workspace\Mejona Cred\AWS\mejona.pem" -o ConnectTimeout=10 -o StrictHostKeyChecking=no ubuntu@65.1.94.25 "journalctl -u contact-service --no-pager -n 20"

echo.
echo 4. Process Check...
ssh -i "D:\Mejona Workspace\Mejona Cred\AWS\mejona.pem" -o ConnectTimeout=10 -o StrictHostKeyChecking=no ubuntu@65.1.94.25 "ps aux | grep contact-service"

echo.
echo 5. Port Check...
ssh -i "D:\Mejona Workspace\Mejona Cred\AWS\mejona.pem" -o ConnectTimeout=10 -o StrictHostKeyChecking=no ubuntu@65.1.94.25 "netstat -tlnp | grep 8081"

echo.
echo Service check completed.
pause