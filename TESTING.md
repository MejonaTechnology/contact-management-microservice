# Contact Management Microservice - Testing Guide

## 📋 Testing Overview

This document provides comprehensive information about the testing strategy, implementation, and execution for the Contact Management Microservice.

## 🧪 Test Types

### 1. Unit Tests
- **Location**: `internal/*/test.go`
- **Purpose**: Test individual components in isolation
- **Coverage**: Handlers, Services, Repository layers
- **Framework**: `testify/suite`, `testify/assert`, `testify/mock`

### 2. Integration Tests
- **Location**: `tests/integration/`
- **Purpose**: Test complete HTTP request/response cycles
- **Coverage**: Full API endpoints with database interactions
- **Framework**: In-memory SQLite database

### 3. Performance Tests
- **Location**: Within unit test files as benchmarks
- **Purpose**: Measure performance characteristics
- **Coverage**: Critical operations and endpoints

## 🏃‍♂️ Running Tests

### Quick Start
```bash
# Run all unit tests
make test

# Run with coverage report
make test-coverage

# Run integration tests
make test-integration

# Run complete test suite
make full-test
```

### Detailed Commands

#### Unit Tests
```bash
# Run all unit tests
make test-unit

# Run specific component tests
make test-handlers    # Handler layer tests
make test-services    # Service layer tests

# Run with race condition detection
make test-race

# Watch mode (requires entr)
make test-watch
```

#### Integration Tests
```bash
# Run integration tests (requires RUN_INTEGRATION_TESTS=true)
make test-integration

# Manual execution
RUN_INTEGRATION_TESTS=true go test ./tests/integration/... -v -tags=integration
```

#### Performance Tests
```bash
# Run benchmarks
make benchmark

# Run specific benchmarks
go test ./internal/services/... -bench=. -benchmem
go test ./internal/handlers/... -bench=. -benchmem
```

#### Coverage Analysis
```bash
# Generate coverage report
make test-coverage

# View coverage in browser
open coverage.html

# Check coverage threshold
go tool cover -func=coverage.out | tail -1
```

## 📊 Test Structure

### Unit Test Organization

#### Handler Tests (`internal/handlers/*_test.go`)
```go
type ContactHandlerTestSuite struct {
    suite.Suite
    handler     *ContactHandler
    mockService *MockContactService
    router      *gin.Engine
}

func (suite *ContactHandlerTestSuite) TestCreateContact_Success() {
    // Arrange: Setup test data and mocks
    // Act: Execute the handler
    // Assert: Verify response and behavior
}
```

#### Service Tests (`internal/services/*_test.go`)
```go
type ContactServiceTestSuite struct {
    suite.Suite
    service  *ContactService
    mockRepo *MockContactRepository
}

func (suite *ContactServiceTestSuite) TestCreateContact_Success() {
    // Test business logic in isolation
}
```

### Integration Test Structure (`tests/integration/`)
```go
type ContactIntegrationTestSuite struct {
    suite.Suite
    db       *gorm.DB
    router   *gin.Engine
    server   *httptest.Server
    authToken string
}

func (suite *ContactIntegrationTestSuite) TestCreateContact_E2E() {
    // Test complete request/response cycle
}
```

## 🎯 Test Coverage Goals

### Current Coverage Targets
- **Overall Coverage**: > 80%
- **Handler Coverage**: > 90%
- **Service Coverage**: > 95%
- **Repository Coverage**: > 85%

### Coverage by Component

#### Handlers (`internal/handlers/`)
- ✅ `contact_test.go` - Contact CRUD operations
- ✅ `health_test.go` - Health check endpoints
- 🚧 `auth_test.go` - Authentication handlers
- 🚧 `analytics_test.go` - Analytics endpoints

#### Services (`internal/services/`)
- ✅ `contact_service_test.go` - Contact business logic
- 🚧 `monitoring_service_test.go` - Monitoring service
- 🚧 `assignment_service_test.go` - Assignment logic

#### Integration (`tests/integration/`)
- ✅ `contact_integration_test.go` - Complete API testing
- 🚧 `auth_integration_test.go` - Authentication flow
- 🚧 `performance_integration_test.go` - Load testing

## 🔧 Test Configuration

### Test Environment Setup
```yaml
# testdata/test_config.yaml
database:
  driver: "sqlite"
  dsn: ":memory:"

server:
  port: "8081"
  host: "localhost"

jwt:
  secret: "test-jwt-secret-key-for-testing-only"
  access_token_duration: "15m"

test:
  fixtures_path: "./testdata/fixtures"
  cleanup_after_tests: true
  parallel_execution: true
  timeout: "30s"
```

### Test Fixtures
- **Contacts**: `testdata/fixtures/contacts.json`
- **Users**: `testdata/fixtures/users.json`
- **Test Data**: Realistic sample data for testing

## 🎨 Testing Patterns

### Mock Usage Pattern
```go
// Create mock
mockService := new(MockContactService)

// Setup expectations
mockService.On("CreateContact", mock.AnythingOfType("*models.Contact")).Return(nil)

// Use in test
handler := &ContactHandler{contactService: mockService}

// Verify expectations
mockService.AssertExpectations(t)
```

### HTTP Testing Pattern
```go
// Setup request
jsonData, _ := json.Marshal(requestData)
req := httptest.NewRequest("POST", "/api/v1/contacts", bytes.NewBuffer(jsonData))
req.Header.Set("Content-Type", "application/json")
w := httptest.NewRecorder()

// Execute
router.ServeHTTP(w, req)

// Assert
assert.Equal(t, http.StatusCreated, w.Code)
```

### Database Testing Pattern
```go
// Setup in-memory database
db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
db.AutoMigrate(&models.Contact{})

// Create test data
contact := &models.Contact{Name: "Test", Email: "test@example.com"}
db.Create(contact)

// Test and cleanup
defer db.Migrator().DropTable(&models.Contact{})
```

## 📈 Performance Benchmarks

### Benchmark Examples
```go
func BenchmarkCreateContact(b *testing.B) {
    service := setupTestService()
    contact := &models.Contact{Name: "Test", Email: "test@example.com"}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.CreateContact(contact)
    }
}
```

### Performance Targets
- **Create Contact**: < 10ms per operation
- **List Contacts**: < 50ms for 100 items
- **Search Contacts**: < 100ms for complex queries
- **Health Check**: < 5ms response time

## 🚨 Test Quality Standards

### Test Naming Convention
```go
func TestMethodName_Scenario_ExpectedBehavior(t *testing.T)
// Examples:
func TestCreateContact_Success_ReturnsCreatedContact(t *testing.T)
func TestGetContact_NotFound_ReturnsError(t *testing.T)
func TestListContacts_WithFilters_ReturnsFilteredResults(t *testing.T)
```

### Test Structure (AAA Pattern)
```go
func TestExample(t *testing.T) {
    // Arrange - Setup test data and dependencies
    contact := &models.Contact{Name: "Test"}
    mockService.On("CreateContact", contact).Return(nil)
    
    // Act - Execute the code under test
    err := service.CreateContact(contact)
    
    // Assert - Verify the expected behavior
    assert.NoError(t, err)
    mockService.AssertExpectations(t)
}
```

### Test Data Management
- Use test fixtures for consistent data
- Clean up after each integration test
- Use in-memory databases for unit tests
- Realistic but minimal test data

## 🛠️ Test Utilities

### Custom Assertions
```go
// Assert HTTP response structure
func AssertAPIResponse(t *testing.T, w *httptest.ResponseRecorder, expectedCode int) {
    assert.Equal(t, expectedCode, w.Code)
    
    var response APIResponse
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.NotEmpty(t, response.Timestamp)
}
```

### Test Helpers
```go
// Create authenticated request
func makeAuthenticatedRequest(method, url string, body interface{}) *http.Response {
    // Implementation details...
}

// Setup test database
func setupTestDatabase() *gorm.DB {
    // Implementation details...
}
```

## 📋 Test Maintenance

### Regular Tasks
- **Weekly**: Review test coverage reports
- **Monthly**: Update test fixtures with new scenarios
- **Quarterly**: Performance benchmark reviews
- **Release**: Full integration test suite execution

### Test Debt Management
- Keep test code quality as high as production code
- Refactor tests when refactoring production code
- Remove obsolete tests promptly
- Update mocks when interfaces change

## 🔍 Debugging Tests

### Common Issues and Solutions

#### Test Flakiness
```bash
# Run tests multiple times to identify flaky tests
for i in {1..10}; do make test || break; done

# Use race detection
make test-race
```

#### Mock Issues
```go
// Always verify mock expectations
defer mockService.AssertExpectations(t)

// Use mock.Anything for flexible matching
mockService.On("Method", mock.Anything).Return(expectedResult)
```

#### Database Issues
```go
// Always clean up after tests
defer func() {
    db.Exec("DELETE FROM contacts")
}()

// Use transactions for isolation
tx := db.Begin()
defer tx.Rollback()
```

## 📊 CI/CD Integration

### GitHub Actions Integration
```yaml
- name: Run Tests
  run: |
    make test-coverage
    make test-integration
    
- name: Upload Coverage
  uses: codecov/codecov-action@v1
  with:
    file: ./coverage.out
```

### Pre-commit Hooks
```bash
#!/bin/sh
# Run tests before commit
make test || exit 1
make test-race || exit 1
```

## 📚 Best Practices

### Do's ✅
- Write tests before or alongside production code
- Use descriptive test names
- Follow the AAA pattern (Arrange, Act, Assert)
- Test both happy path and error cases
- Use mocks to isolate units under test
- Keep tests fast and independent
- Use table-driven tests for multiple scenarios

### Don'ts ❌
- Don't test implementation details
- Don't create tests that depend on external services
- Don't ignore failing tests
- Don't write tests that test the framework
- Don't use real databases in unit tests
- Don't create overly complex test setups

## 🎯 Next Steps

### Planned Improvements
1. **Enhanced Integration Tests**
   - Authentication flow testing
   - Error handling scenarios
   - Rate limiting validation

2. **Performance Testing**
   - Load testing with realistic data volumes
   - Stress testing for concurrent requests
   - Memory usage profiling

3. **Contract Testing**
   - API contract validation
   - Backward compatibility testing
   - Client SDK integration tests

4. **Security Testing**
   - Input validation testing
   - SQL injection prevention
   - Authentication bypass attempts

---

## 📞 Support

### Getting Help
- **Documentation**: This testing guide
- **Code Examples**: See test files for patterns
- **Issues**: Report testing issues via GitHub Issues
- **Team Chat**: Discuss testing strategies in team channels

### Contributing to Tests
1. Follow existing patterns and conventions
2. Ensure new features include comprehensive tests
3. Update this documentation when adding new test types
4. Review test coverage before submitting PRs

---

**© 2024 Mejona Technology LLP. All rights reserved.**

*This testing guide ensures comprehensive coverage and quality assurance for the Contact Management Microservice.*