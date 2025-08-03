@echo off
echo ===============================================
echo TESTING ALL 20 CONTACT SERVICE ENDPOINTS
echo ===============================================
echo.

echo === HEALTH ENDPOINTS (6) ===
echo 1. Basic Health Check:
curl -s http://localhost:8081/health
echo.
echo.

echo 2. Deep Health Check:
curl -s http://localhost:8081/health/deep
echo.
echo.

echo 3. Status Check:
curl -s http://localhost:8081/status
echo.
echo.

echo 4. Readiness Check:
curl -s http://localhost:8081/ready
echo.
echo.

echo 5. Liveness Check:
curl -s http://localhost:8081/alive
echo.
echo.

echo 6. Metrics Check:
curl -s http://localhost:8081/metrics
echo.
echo.

echo === DASHBOARD ENDPOINTS (7) ===
echo 7. List Contacts:
curl -s http://localhost:8081/api/v1/dashboard/contacts
echo.
echo.

echo 8. Contact Statistics:
curl -s http://localhost:8081/api/v1/dashboard/contacts/stats
echo.
echo.

echo 9. Get Contact by ID:
curl -s http://localhost:8081/api/v1/dashboard/contacts/1
echo.
echo.

echo 10. Create Contact:
curl -s -X POST -H "Content-Type: application/json" -d "{\"name\":\"Test All\",\"email\":\"testall@endpoint.com\",\"message\":\"Testing all endpoints\"}" http://localhost:8081/api/v1/dashboard/contact
echo.
echo.

echo 11. Update Contact Status:
curl -s -X PUT -H "Content-Type: application/json" -d "{\"status\":\"resolved\"}" http://localhost:8081/api/v1/dashboard/contacts/1/status
echo.
echo.

echo 12. Export Contacts:
curl -s http://localhost:8081/api/v1/dashboard/contacts/export
echo.
echo.

echo 13. Bulk Update:
curl -s -X POST -H "Content-Type: application/json" -d "{\"ids\":[1,2],\"status\":\"new\"}" http://localhost:8081/api/v1/dashboard/contacts/bulk-update
echo.
echo.

echo === AUTH ENDPOINTS (6) ===
echo 14. Login:
curl -s -X POST -H "Content-Type: application/json" -d "{\"email\":\"admin@mejona.com\",\"password\":\"admin123\"}" http://localhost:8081/api/v1/auth/login
echo.
echo.

echo 15. Refresh Token:
curl -s -X POST -H "Content-Type: application/json" -d "{\"token\":\"simple-test-token\"}" http://localhost:8081/api/v1/auth/refresh
echo.
echo.

echo 16. Get Profile:
curl -s -H "Authorization: Bearer simple-test-token" http://localhost:8081/api/v1/auth/profile
echo.
echo.

echo 17. Validate Token:
curl -s -H "Authorization: Bearer simple-test-token" http://localhost:8081/api/v1/auth/validate
echo.
echo.

echo 18. Logout:
curl -s -X POST -H "Authorization: Bearer simple-test-token" http://localhost:8081/api/v1/auth/logout
echo.
echo.

echo 19. Change Password:
curl -s -X POST -H "Authorization: Bearer simple-test-token" -H "Content-Type: application/json" -d "{\"old_password\":\"admin123\",\"new_password\":\"newpass123\"}" http://localhost:8081/api/v1/auth/change-password
echo.
echo.

echo === OTHER ENDPOINTS (1) ===
echo 20. Test Endpoint:
curl -s http://localhost:8081/api/v1/test
echo.
echo.

echo 21. Public Contact:
curl -s -X POST -H "Content-Type: application/json" -d "{\"name\":\"Public Test\",\"email\":\"public@test.com\",\"message\":\"Testing public endpoint\"}" http://localhost:8081/api/v1/public/contact
echo.
echo.

echo ===============================================
echo ENDPOINT TESTING COMPLETE
echo ===============================================