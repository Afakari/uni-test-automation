#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_godog() {
    if ! command -v godog &> /dev/null; then
        print_error "godog is not installed. Please install it first:"
        echo "  go install github.com/cucumber/godog/cmd/godog@latest"
        exit 1
    fi
}

check_go_mod() {
    if [ ! -f "go.mod" ]; then
        print_error "go.mod not found. Please run this script from the go project root."
        exit 1
    fi
}

main() {
    print_status "Starting BDD Test Runner..."

    check_go_mod
    check_godog

    print_status "Running all BDD tests via main_test.go..."

    if go test -v ./internal/features/steps/; then
        print_success "All tests passed"
    else
        print_error "Tests failed"
        exit 1
    fi
}

main