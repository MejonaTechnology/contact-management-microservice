#!/bin/bash

# =================================================================
# Contact Management Microservice - Deployment Verification Script
# =================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_URL="${1:-http://localhost:8081}"
TOTAL_ENDPOINTS=20
PASSED_TESTS=0

echo -e "${BLUE}🚀 Contact Management Microservice - Deployment Verification${NC}"
echo -e "${BLUE}=================================================================${NC}"
echo "Service URL: $SERVICE_URL"
echo "Testing $TOTAL_ENDPOINTS endpoints..."
echo ""

# Function to test endpoint
test_endpoint() {
    local method="$1"
    local endpoint="$2"
    local description="$3"
    local expected_status="${4:-200}"
    local additional_args="$5"
    
    echo -n "Testing $method $endpoint ($description)... "
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -o /dev/null -w "%{http_code}" $additional_args "$SERVICE_URL$endpoint")
    else
        response=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" $additional_args "$SERVICE_URL$endpoint")
    fi
    
    if [ "$response" = "$expected_status" ]; then
        echo -e "${GREEN}✅ PASS${NC} (HTTP $response)"
        ((PASSED_TESTS++))
        return 0
    else
        echo -e "${RED}❌ FAIL${NC} (HTTP $response, expected $expected_status)"
        return 1
    fi
}

# Function to test endpoint with auth
test_auth_endpoint() {
    local method="$1"
    local endpoint="$2"
    local description="$3"
    local expected_status="${4:-401}"
    
    test_endpoint "$method" "$endpoint" "$description (no auth)" "$expected_status"
}

echo -e "${YELLOW}📋 HEALTH & MONITORING ENDPOINTS (6 endpoints)${NC}"
echo "================================================"

# 1. Basic Health Check
test_endpoint "GET" "/health" "Basic Health Check"

# 2. Detailed Health Check
test_endpoint "GET" "/health/detailed" "Detailed Health Check"

# 3. Readiness Check
test_endpoint "GET" "/health/ready" "Readiness Check"

# 4. Liveness Check
test_endpoint "GET" "/health/live" "Liveness Check"

# 5. Metrics Endpoint
test_endpoint "GET" "/metrics" "Prometheus Metrics"

# 6. System Info
test_endpoint "GET" "/health/info" "System Information"

echo ""
echo -e "${YELLOW}🔐 AUTHENTICATION ENDPOINTS (2 endpoints)${NC}"
echo "=========================================="

# 7. Login Endpoint (POST with invalid data should return 400/401)
test_endpoint "POST" "/api/v1/auth/login" "Login Endpoint" "400" '-H "Content-Type: application/json" -d "{}"'

# 8. Token Validation (without token should return 401)
test_auth_endpoint "GET" "/api/v1/auth/validate" "Token Validation"

echo ""
echo -e "${YELLOW}📞 CONTACT MANAGEMENT ENDPOINTS (5 endpoints)${NC}"
echo "=============================================="

# 9. Get All Contacts (requires auth)
test_auth_endpoint "GET" "/api/v1/contacts" "Get All Contacts"

# 10. Create Contact (requires auth)
test_auth_endpoint "POST" "/api/v1/contacts" "Create Contact"

# 11. Get Contact by ID (requires auth)
test_auth_endpoint "GET" "/api/v1/contacts/1" "Get Contact by ID"

# 12. Update Contact (requires auth)
test_auth_endpoint "PUT" "/api/v1/contacts/1" "Update Contact"

# 13. Delete Contact (requires auth)
test_auth_endpoint "DELETE" "/api/v1/contacts/1" "Delete Contact"

echo ""
echo -e "${YELLOW}🔍 SEARCH & FILTER ENDPOINTS (2 endpoints)${NC}"
echo "==========================================="

# 14. Search Contacts (requires auth)
test_auth_endpoint "GET" "/api/v1/contacts/search?q=test" "Search Contacts"

# 15. Advanced Search (requires auth)
test_auth_endpoint "POST" "/api/v1/contacts/search/advanced" "Advanced Search"

echo ""
echo -e "${YELLOW}📊 ANALYTICS & REPORTING ENDPOINTS (2 endpoints)${NC}"
echo "==============================================="

# 16. Contact Analytics (requires auth)
test_auth_endpoint "GET" "/api/v1/analytics/contacts" "Contact Analytics"

# 17. Dashboard Stats (requires auth)
test_auth_endpoint "GET" "/api/v1/analytics/dashboard" "Dashboard Statistics"

echo ""
echo -e "${YELLOW}⚙️ CONFIGURATION ENDPOINTS (3 endpoints)${NC}"
echo "========================================="

# 18. Contact Types (requires auth)
test_auth_endpoint "GET" "/api/v1/contact-types" "Contact Types"

# 19. Contact Sources (requires auth)
test_auth_endpoint "GET" "/api/v1/contact-sources" "Contact Sources"

# 20. Export Contacts (requires auth)
test_auth_endpoint "GET" "/api/v1/contacts/export" "Export Contacts"

echo ""
echo -e "${BLUE}=================================================================${NC}"
echo -e "${BLUE}📊 DEPLOYMENT VERIFICATION SUMMARY${NC}"
echo -e "${BLUE}=================================================================${NC}"

if [ $PASSED_TESTS -eq $TOTAL_ENDPOINTS ]; then
    echo -e "${GREEN}🎉 SUCCESS: All $TOTAL_ENDPOINTS endpoints are responding correctly!${NC}"
    echo -e "${GREEN}✅ Contact Management Microservice is properly deployed${NC}"
    echo ""
    echo -e "${YELLOW}📋 Service Information:${NC}"
    echo "   • Service URL: $SERVICE_URL"
    echo "   • Health Check: $SERVICE_URL/health"
    echo "   • API Documentation: $SERVICE_URL/swagger/index.html"
    echo "   • Metrics: $SERVICE_URL/metrics"
    echo ""
    echo -e "${YELLOW}🔗 Next Steps:${NC}"
    echo "   1. Configure authentication with valid JWT tokens"
    echo "   2. Set up monitoring and alerting"
    echo "   3. Configure SSL certificate for HTTPS"
    echo "   4. Set up automated backups"
    echo ""
    exit 0
else
    echo -e "${RED}❌ FAILURE: $PASSED_TESTS/$TOTAL_ENDPOINTS endpoints passed${NC}"
    echo -e "${RED}⚠️  Some endpoints are not responding correctly${NC}"
    echo ""
    echo -e "${YELLOW}🔧 Troubleshooting Steps:${NC}"
    echo "   1. Check service status: sudo systemctl status contact-service"
    echo "   2. Check service logs: sudo journalctl -u contact-service -f"
    echo "   3. Check application logs: tail -f /opt/mejona/contact-management-microservice/logs/contact-service.log"
    echo "   4. Verify database connectivity"
    echo "   5. Check environment configuration"
    echo ""
    exit 1
fi