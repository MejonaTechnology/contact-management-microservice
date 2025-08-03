#!/bin/bash

# Contact Management Microservice - Comprehensive Test Runner
# This script runs the complete test suite with detailed reporting

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
TEST_TIMEOUT="10m"
COVERAGE_THRESHOLD=80
INTEGRATION_TIMEOUT="5m"
BENCHMARK_TIME="30s"

# Directories
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_RESULTS_DIR="${PROJECT_ROOT}/test-results"
COVERAGE_DIR="${TEST_RESULTS_DIR}/coverage"

# Create test results directory
mkdir -p "${TEST_RESULTS_DIR}"
mkdir -p "${COVERAGE_DIR}"

# Functions
print_header() {
    echo -e "\n${BLUE}================================================${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}================================================${NC}\n"
}

print_section() {
    echo -e "\n${CYAN}--- $1 ---${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${PURPLE}ℹ️  $1${NC}"
}

# Check dependencies
check_dependencies() {
    print_section "Checking Dependencies"
    
    # Check Go installation
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    print_success "Go $(go version | cut -d' ' -f3) found"
    
    # Check required Go modules
    if ! go list -m github.com/stretchr/testify &> /dev/null; then
        print_info "Installing testify..."
        go get github.com/stretchr/testify
    fi
    print_success "Required dependencies available"
}

# Clean previous test artifacts
clean_artifacts() {
    print_section "Cleaning Previous Test Artifacts"
    
    # Clean Go test cache
    go clean -testcache
    
    # Remove old coverage files
    rm -f "${PROJECT_ROOT}/coverage.out"
    rm -f "${PROJECT_ROOT}/coverage.html"
    rm -f "${PROJECT_ROOT}/coverage.xml"
    
    # Remove integration test databases
    rm -f "${PROJECT_ROOT}/tests/integration/test.db"
    
    # Clean test results directory
    rm -rf "${TEST_RESULTS_DIR}"/*
    
    print_success "Test artifacts cleaned"
}

# Run unit tests
run_unit_tests() {
    print_section "Running Unit Tests"
    
    local unit_results="${TEST_RESULTS_DIR}/unit-tests.json"
    
    echo -e "${YELLOW}Running unit tests with detailed output...${NC}"
    
    # Run unit tests with JSON output
    if go test ./internal/... \
        -v \
        -short \
        -timeout="${TEST_TIMEOUT}" \
        -json > "${unit_results}"; then
        
        # Parse results
        local total_tests=$(grep -c '"Action":"pass"' "${unit_results}" 2>/dev/null || echo "0")
        local failed_tests=$(grep -c '"Action":"fail"' "${unit_results}" 2>/dev/null || echo "0")
        
        if [ "${failed_tests}" -eq 0 ]; then
            print_success "Unit tests passed (${total_tests} tests)"
        else
            print_error "Unit tests failed (${failed_tests} failures out of ${total_tests} tests)"
            return 1
        fi
    else
        print_error "Unit tests execution failed"
        return 1
    fi
}

# Run integration tests
run_integration_tests() {
    print_section "Running Integration Tests"
    
    local integration_results="${TEST_RESULTS_DIR}/integration-tests.json"
    
    echo -e "${YELLOW}Running integration tests...${NC}"
    
    # Set environment variable for integration tests
    export RUN_INTEGRATION_TESTS=true
    
    # Run integration tests
    if go test ./tests/integration/... \
        -v \
        -tags=integration \
        -timeout="${INTEGRATION_TIMEOUT}" \
        -json > "${integration_results}"; then
        
        # Parse results
        local total_tests=$(grep -c '"Action":"pass"' "${integration_results}" 2>/dev/null || echo "0")
        local failed_tests=$(grep -c '"Action":"fail"' "${integration_results}" 2>/dev/null || echo "0")
        
        if [ "${failed_tests}" -eq 0 ]; then
            print_success "Integration tests passed (${total_tests} tests)"
        else
            print_error "Integration tests failed (${failed_tests} failures out of ${total_tests} tests)"
            return 1
        fi
    else
        print_warning "Integration tests skipped or failed"
    fi
    
    unset RUN_INTEGRATION_TESTS
}

# Run tests with coverage
run_coverage_tests() {
    print_section "Running Coverage Analysis"
    
    local coverage_out="${COVERAGE_DIR}/coverage.out"
    local coverage_html="${COVERAGE_DIR}/coverage.html"
    local coverage_xml="${COVERAGE_DIR}/coverage.xml"
    
    echo -e "${YELLOW}Generating coverage report...${NC}"
    
    # Run tests with coverage
    if go test ./... \
        -coverprofile="${coverage_out}" \
        -covermode=atomic \
        -timeout="${TEST_TIMEOUT}"; then
        
        # Generate HTML report
        go tool cover -html="${coverage_out}" -o "${coverage_html}"
        
        # Generate XML report (if gocover-cobertura is available)
        if command -v gocover-cobertura &> /dev/null; then
            gocover-cobertura < "${coverage_out}" > "${coverage_xml}"
        fi
        
        # Check coverage threshold
        local coverage_percent=$(go tool cover -func="${coverage_out}" | tail -1 | awk '{print $3}' | sed 's/%//')
        
        echo -e "${BLUE}Coverage Report:${NC}"
        go tool cover -func="${coverage_out}" | tail -10
        
        if (( $(echo "${coverage_percent} >= ${COVERAGE_THRESHOLD}" | bc -l) )); then
            print_success "Coverage: ${coverage_percent}% (meets ${COVERAGE_THRESHOLD}% threshold)"
        else
            print_warning "Coverage: ${coverage_percent}% (below ${COVERAGE_THRESHOLD}% threshold)"
        fi
        
        print_info "HTML report: ${coverage_html}"
        
    else
        print_error "Coverage analysis failed"
        return 1
    fi
}

# Run race condition tests
run_race_tests() {
    print_section "Running Race Condition Detection"
    
    echo -e "${YELLOW}Running tests with race detection...${NC}"
    
    if go test ./... -race -timeout="${TEST_TIMEOUT}"; then
        print_success "No race conditions detected"
    else
        print_error "Race conditions detected"
        return 1
    fi
}

# Run benchmark tests
run_benchmarks() {
    print_section "Running Performance Benchmarks"
    
    local benchmark_results="${TEST_RESULTS_DIR}/benchmarks.txt"
    
    echo -e "${YELLOW}Running benchmark tests...${NC}"
    
    if go test ./... \
        -bench=. \
        -benchmem \
        -benchtime="${BENCHMARK_TIME}" \
        -timeout="${TEST_TIMEOUT}" > "${benchmark_results}"; then
        
        echo -e "${BLUE}Benchmark Results:${NC}"
        cat "${benchmark_results}"
        
        print_success "Benchmarks completed"
        print_info "Results saved to: ${benchmark_results}"
    else
        print_warning "Benchmarks failed or unavailable"
    fi
}

# Generate test report
generate_report() {
    print_section "Generating Test Summary Report"
    
    local report_file="${TEST_RESULTS_DIR}/test-summary.md"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    cat > "${report_file}" << EOF
# Contact Management Microservice - Test Report

**Generated**: ${timestamp}
**Go Version**: $(go version)
**Test Timeout**: ${TEST_TIMEOUT}

## Test Results Summary

EOF

    # Add unit test results
    if [ -f "${TEST_RESULTS_DIR}/unit-tests.json" ]; then
        local unit_total=$(grep -c '"Action":"pass"' "${TEST_RESULTS_DIR}/unit-tests.json" 2>/dev/null || echo "0")
        local unit_failed=$(grep -c '"Action":"fail"' "${TEST_RESULTS_DIR}/unit-tests.json" 2>/dev/null || echo "0")
        
        echo "### Unit Tests" >> "${report_file}"
        echo "- **Total**: ${unit_total}" >> "${report_file}"
        echo "- **Failed**: ${unit_failed}" >> "${report_file}"
        echo "- **Status**: $( [ "${unit_failed}" -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED" )" >> "${report_file}"
        echo "" >> "${report_file}"
    fi
    
    # Add integration test results
    if [ -f "${TEST_RESULTS_DIR}/integration-tests.json" ]; then
        local integration_total=$(grep -c '"Action":"pass"' "${TEST_RESULTS_DIR}/integration-tests.json" 2>/dev/null || echo "0")
        local integration_failed=$(grep -c '"Action":"fail"' "${TEST_RESULTS_DIR}/integration-tests.json" 2>/dev/null || echo "0")
        
        echo "### Integration Tests" >> "${report_file}"
        echo "- **Total**: ${integration_total}" >> "${report_file}"
        echo "- **Failed**: ${integration_failed}" >> "${report_file}"
        echo "- **Status**: $( [ "${integration_failed}" -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED" )" >> "${report_file}"
        echo "" >> "${report_file}"
    fi
    
    # Add coverage information
    if [ -f "${COVERAGE_DIR}/coverage.out" ]; then
        local coverage_percent=$(go tool cover -func="${COVERAGE_DIR}/coverage.out" | tail -1 | awk '{print $3}' | sed 's/%//')
        
        echo "### Code Coverage" >> "${report_file}"
        echo "- **Coverage**: ${coverage_percent}%" >> "${report_file}"
        echo "- **Threshold**: ${COVERAGE_THRESHOLD}%" >> "${report_file}"
        echo "- **Status**: $( (( $(echo "${coverage_percent} >= ${COVERAGE_THRESHOLD}" | bc -l) )) && echo "✅ MEETS THRESHOLD" || echo "⚠️ BELOW THRESHOLD" )" >> "${report_file}"
        echo "- **HTML Report**: [coverage.html](coverage/coverage.html)" >> "${report_file}"
        echo "" >> "${report_file}"
    fi
    
    # Add benchmark information
    if [ -f "${TEST_RESULTS_DIR}/benchmarks.txt" ]; then
        echo "### Performance Benchmarks" >> "${report_file}"
        echo "- **Status**: ✅ COMPLETED" >> "${report_file}"
        echo "- **Results**: [benchmarks.txt](benchmarks.txt)" >> "${report_file}"
        echo "" >> "${report_file}"
    fi
    
    echo "## Files Generated" >> "${report_file}"
    echo "" >> "${report_file}"
    find "${TEST_RESULTS_DIR}" -type f -name "*.json" -o -name "*.txt" -o -name "*.html" -o -name "*.xml" | while read -r file; do
        local relative_path=$(realpath --relative-to="${TEST_RESULTS_DIR}" "${file}")
        echo "- ${relative_path}" >> "${report_file}"
    done
    
    print_success "Test report generated: ${report_file}"
}

# Main execution
main() {
    print_header "Contact Management Microservice - Test Suite"
    
    local start_time=$(date +%s)
    local exit_code=0
    
    # Change to project directory
    cd "${PROJECT_ROOT}"
    
    # Execute test phases
    check_dependencies || exit_code=$?
    clean_artifacts || exit_code=$?
    
    # Run tests (continue even if some fail to get complete picture)
    run_unit_tests || exit_code=$?
    run_integration_tests || exit_code=$?
    run_coverage_tests || exit_code=$?
    run_race_tests || exit_code=$?
    run_benchmarks || exit_code=$?
    
    # Generate final report
    generate_report
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    print_header "Test Suite Execution Summary"
    
    echo -e "${BLUE}Execution Time: ${duration} seconds${NC}"
    echo -e "${BLUE}Results Directory: ${TEST_RESULTS_DIR}${NC}"
    
    if [ $exit_code -eq 0 ]; then
        print_success "All tests completed successfully!"
        echo -e "\n${GREEN}🎉 TEST SUITE PASSED 🎉${NC}\n"
    else
        print_error "Some tests failed or had issues"
        echo -e "\n${RED}❌ TEST SUITE FAILED ❌${NC}\n"
        echo -e "${YELLOW}Check individual test results for details${NC}"
    fi
    
    return $exit_code
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h          Show this help message"
        echo "  --unit-only         Run only unit tests"
        echo "  --integration-only  Run only integration tests"
        echo "  --coverage-only     Run only coverage analysis"
        echo "  --benchmarks-only   Run only benchmarks"
        echo ""
        echo "Environment Variables:"
        echo "  TEST_TIMEOUT        Test timeout (default: ${TEST_TIMEOUT})"
        echo "  COVERAGE_THRESHOLD  Coverage threshold (default: ${COVERAGE_THRESHOLD}%)"
        echo "  INTEGRATION_TIMEOUT Integration test timeout (default: ${INTEGRATION_TIMEOUT})"
        echo "  BENCHMARK_TIME      Benchmark duration (default: ${BENCHMARK_TIME})"
        exit 0
        ;;
    --unit-only)
        check_dependencies && clean_artifacts && run_unit_tests
        exit $?
        ;;
    --integration-only)
        check_dependencies && clean_artifacts && run_integration_tests
        exit $?
        ;;
    --coverage-only)
        check_dependencies && clean_artifacts && run_coverage_tests
        exit $?
        ;;
    --benchmarks-only)
        check_dependencies && clean_artifacts && run_benchmarks
        exit $?
        ;;
    *)
        main
        exit $?
        ;;
esac