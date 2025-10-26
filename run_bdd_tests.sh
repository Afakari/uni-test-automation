#!/bin/bash

# BDD Test Runner for Go Todo Application
# This script runs all BDD tests using godog with proper configuration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
JWT_SECRET=${JWT_SECRET:-"test-secret"}
TEST_TIMEOUT=${TEST_TIMEOUT:-"30s"}
TEST_RETRY_COUNT=${TEST_RETRY_COUNT:-3}
TEST_LOG_LEVEL=${TEST_LOG_LEVEL:-"info"}
TEST_CLEANUP_AFTER=${TEST_CLEANUP_AFTER:-true}
TEST_PARALLEL=${TEST_PARALLEL:-false}
GODOG_TAGS=${GODOG_TAGS:-""}

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}
print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if godog is installed
check_godog() {
    if ! command -v godog &> /dev/null; then
        print_error "godog is not installed. Please install it first:"
        echo "  go install github.com/cucumber/godog/cmd/godog@latest"
        exit 1
    fi
}

# Function to check if go mod is available
check_go_mod() {
    if [ ! -f "go.mod" ]; then
        print_error "go.mod not found. Please run this script from the go project root."
        exit 1
    fi
}

# Function to run a specific test suite
run_test_suite() {
    local feature_name=$1
    local feature_file=$2
    local test_function=$3
    
    print_status "Running $feature_name tests..."
    
    if [ ! -f "$feature_file" ]; then
        print_warning "Feature file $feature_file not found, skipping..."
        return 0
    fi
    
    # Set environment variables for this test run
    export JWT_SECRET="$JWT_SECRET"
    export TEST_TIMEOUT="$TEST_TIMEOUT"
    export TEST_RETRY_COUNT="$TEST_RETRY_COUNT"
    export TEST_LOG_LEVEL="$TEST_LOG_LEVEL"
    export TEST_CLEANUP_AFTER="$TEST_CLEANUP_AFTER"
    export TEST_PARALLEL="$TEST_PARALLEL"
    
    # Build the test command
    local test_cmd="go test -v -run $test_function ./internal/features/steps/"
    
    # Add tags if specified
    if [ -n "$GODOG_TAGS" ]; then
        test_cmd="$test_cmd -godog.tags=$GODOG_TAGS"
    fi
    
    # Run the test
    if eval "$test_cmd"; then
        print_success "$feature_name tests passed"
        return 0
    else
        print_error "$feature_name tests failed"
        return 1
    fi
}

# Function to run all tests
run_all_tests() {
    print_status "Running all BDD tests..."
    
    local failed_tests=()
    local passed_tests=()
    
    # Define test suites
    local test_suites=(
        "Authentication:internal/features/authentication.feature:TestUserAuthentication"
        "Authorization:internal/features/authorization.feature:TestUserAuthorization"
        "TodoManagement:internal/features/todo_management.feature:TestTodoManagement"
        "ConcurrentUpdate:internal/features/concurrent_update.feature:TestConcurrentUpdate"
        "Integration:internal/features/integration.feature:TestIntegration"
    )
    
    # Run each test suite
    for suite in "${test_suites[@]}"; do
        IFS=':' read -r name feature_file test_function <<< "$suite"
        
        if run_test_suite "$name" "$feature_file" "$test_function"; then
            passed_tests+=("$name")
        else
            failed_tests+=("$name")
        fi
    done
    
    # Print summary
    echo ""
    print_status "Test Summary:"
    print_success "Passed: ${#passed_tests[@]} test suites"
    for test in "${passed_tests[@]}"; do
        echo "  ✓ $test"
    done
    
    if [ ${#failed_tests[@]} -gt 0 ]; then
        print_error "Failed: ${#failed_tests[@]} test suites"
        for test in "${failed_tests[@]}"; do
            echo "  ✗ $test"
        done
        return 1
    fi
    
    return 0
}

# Function to run tests with specific tags
run_tagged_tests() {
    local tags=$1
    print_status "Running tests with tags: $tags"
    
    export JWT_SECRET="$JWT_SECRET"
    export TEST_TIMEOUT="$TEST_TIMEOUT"
    export TEST_RETRY_COUNT="$TEST_RETRY_COUNT"
    export TEST_LOG_LEVEL="$TEST_LOG_LEVEL"
    export TEST_CLEANUP_AFTER="$TEST_CLEANUP_AFTER"
    export TEST_PARALLEL="$TEST_PARALLEL"
    
    go test -v -godog.tags="$tags" ./internal/features/steps/
}

# Function to run tests in parallel
run_parallel_tests() {
    print_status "Running tests in parallel..."
    
    export JWT_SECRET="$JWT_SECRET"
    export TEST_TIMEOUT="$TEST_TIMEOUT"
    export TEST_RETRY_COUNT="$TEST_RETRY_COUNT"
    export TEST_LOG_LEVEL="$TEST_LOG_LEVEL"
    export TEST_CLEANUP_AFTER="$TEST_CLEANUP_AFTER"
    export TEST_PARALLEL="true"
    
    go test -v -parallel 4  ./internal/features/steps/
}

# Function to generate test report
generate_report() {
    local report_dir="test-reports"
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local report_file="$report_dir/bdd_test_report_$timestamp.html"
    
    print_status "Generating test report..."
    
    mkdir -p "$report_dir"
    
    export JWT_SECRET="$JWT_SECRET"
    export TEST_TIMEOUT="$TEST_TIMEOUT"
    export TEST_RETRY_COUNT="$TEST_RETRY_COUNT"
    export TEST_LOG_LEVEL="$TEST_LOG_LEVEL"
    export TEST_CLEANUP_AFTER="$TEST_CLEANUP_AFTER"
    export TEST_PARALLEL="$TEST_PARALLEL"
    
    go test -v -godog.output="$report_file" ./internal/features/steps/
    
    print_success "Test report generated: $report_file"
}

# Function to show help
show_help() {
    echo "BDD Test Runner for Go Todo Application"
    echo ""
    echo "Usage: $0 [OPTIONS] [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  all                    Run all BDD tests (default)"
    echo "  auth                   Run authentication tests only"
    echo "  todo                   Run todo management tests only"
    echo "  integration            Run integration tests only"
    echo "  concurrent             Run concurrent update tests only"
    echo "  tags <tags>            Run tests with specific tags"
    echo "  parallel               Run tests in parallel"
    echo "  report                 Generate test report"
    echo "  help                   Show this help message"
    echo ""
    echo "Options:"
    echo "  --jwt-secret <secret>  Set JWT secret (default: test-secret)"
    echo "  --timeout <duration>   Set test timeout (default: 30s)"
    echo "  --retry-count <count>  Set retry count (default: 3)"
    echo "  --log-level <level>    Set log level (default: info)"
    echo "  --no-cleanup          Disable cleanup after tests"
    echo "  --parallel             Enable parallel test execution"
    echo ""
    echo "Environment Variables:"
    echo "  JWT_SECRET             JWT secret for tests"
    echo "  TEST_TIMEOUT           Test timeout duration"
    echo "  TEST_RETRY_COUNT       Number of retries for failed tests"
    echo "  TEST_LOG_LEVEL         Log level for tests"
    echo "  TEST_CLEANUP_AFTER     Cleanup after tests (true/false)"
    echo "  TEST_PARALLEL          Run tests in parallel (true/false)"
    echo "  GODOG_TAGS             Godog tags to filter tests"
    echo ""
    echo "Examples:"
    echo "  $0                     # Run all tests"
    echo "  $0 auth                # Run authentication tests only"
    echo "  $0 tags @smoke         # Run tests tagged with @smoke"
    echo "  $0 --format progress   # Run with progress format"
    echo "  $0 --parallel          # Run tests in parallel"
}

# Parse command line arguments
COMMAND="all"
while [[ $# -gt 0 ]]; do
    case $1 in
        --jwt-secret)
            JWT_SECRET="$2"
            shift 2
            ;;
        --timeout)
            TEST_TIMEOUT="$2"
            shift 2
            ;;
        --retry-count)
            TEST_RETRY_COUNT="$2"
            shift 2
            ;;
        --log-level)
            TEST_LOG_LEVEL="$2"
            shift 2
            ;;
        --no-cleanup)
            TEST_CLEANUP_AFTER="false"
            shift
            ;;
        --parallel)
            TEST_PARALLEL="true"
            shift
            ;;
        all|auth|todo|integration|concurrent|tags|parallel|report|help)
            COMMAND="$1"
            if [ "$1" = "tags" ]; then
                GODOG_TAGS="$2"
                shift
            fi
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Main execution
main() {
    print_status "Starting BDD Test Runner..."
    print_status "Configuration:"
    echo "  JWT Secret: $JWT_SECRET"
    echo "  Timeout: $TEST_TIMEOUT"
    echo "  Retry Count: $TEST_RETRY_COUNT"
    echo "  Log Level: $TEST_LOG_LEVEL"
    echo "  Cleanup After: $TEST_CLEANUP_AFTER"
    echo "  Parallel: $TEST_PARALLEL"
    if [ -n "$GODOG_TAGS" ]; then
        echo "  Tags: $GODOG_TAGS"
    fi
    echo ""
    
    # Check prerequisites
    check_go_mod
    check_godog
    
    # Execute based on command
    case $COMMAND in
        all)
            run_all_tests
            ;;
        auth)
            run_test_suite "Authentication" "internal/features/authentication.feature" "TestUserAuthentication"
            ;;
        todo)
            run_test_suite "TodoManagement" "internal/features/todo_management.feature" "TestTodoManagement"
            ;;
        integration)
            run_test_suite "Integration" "internal/features/integration.feature" "TestIntegration"
            ;;
        concurrent)
            run_test_suite "ConcurrentUpdate" "internal/features/concurrent_update.feature" "TestConcurrentUpdate"
            ;;
        tags)
            run_tagged_tests "$GODOG_TAGS"
            ;;
        parallel)
            run_parallel_tests
            ;;
        report)
            generate_report
            ;;
        help)
            show_help
            ;;
        *)
            print_error "Unknown command: $COMMAND"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
