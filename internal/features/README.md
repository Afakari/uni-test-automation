# Gherkin Test Scenarios for REST API

This directory contains comprehensive Gherkin feature files for testing the Todo REST API. The scenarios are organized by functionality and cover all aspects of the API.

## Feature Files Overview

### 1. `authentication.feature`
Tests user registration and login functionality:
- User registration with valid/invalid credentials
- Login with valid/invalid credentials
- Error handling for authentication failures
- JSON validation and error responses

### 2. `todo_management.feature`
Tests CRUD operations for todos:
- Create, read, update, and delete todos
- Authentication requirements for protected endpoints
- Data validation and error handling
- Complete todo lifecycle management

### 3. `authorization.feature`
Tests user data isolation and security:
- Users can only access their own todos
- Cross-user data access prevention
- Token validation and authentication middleware
- Authorization error handling

### 4. `error_handling.feature`
Tests error scenarios and edge cases:
- Invalid JSON handling
- Missing authentication
- Server errors and timeouts
- Special characters and edge cases
- Malformed requests

### 5. `performance.feature`
Tests system performance under load:
- Concurrent user operations
- High volume data operations
- Memory usage and response times
- Load testing scenarios

### 6. `integration.feature`
Tests end-to-end user workflows:
- Complete user journeys from registration to todo management
- Multi-user scenarios with data isolation
- Session management and token handling
- Data persistence and consistency

### 7. `concurrent_update.feature`
Tests race conditions and concurrent operations:
- Simultaneous updates to the same todo
- Data loss scenarios
- Unpredictable behavior demonstration
- Concurrent delete and update operations

## Test Execution

### Prerequisites
- Go environment with required dependencies
- JWT secret configured
- Test database/storage setup
- Gherkin test runner (e.g., godog)

### Running Tests
```bash
# Run all feature tests
godog

# Run specific feature
godog features/authentication.feature

# Run with verbose output
godog -v

# Run with specific tags
godog --tags @smoke
```

### Test Data Setup
Each feature file includes background steps that set up the necessary test data:
- Secret key configuration
- User registration and authentication
- Test todo creation
- Environment preparation

## Test Coverage

### Functional Coverage
- ✅ User authentication (registration/login)
- ✅ Todo CRUD operations
- ✅ User authorization and data isolation
- ✅ Error handling and edge cases
- ✅ Performance and load testing
- ✅ Integration and end-to-end workflows
- ✅ Race conditions and concurrent operations

### Non-Functional Coverage
- ✅ Security testing (authentication/authorization)
- ✅ Performance testing (load, memory, response times)
- ✅ Reliability testing (error handling, recovery)
- ✅ Usability testing (user workflows)

## Test Scenarios Summary

| Feature | Scenarios | Focus Area |
|---------|-----------|------------|
| Authentication | 9 scenarios | User registration, login, error handling |
| Todo Management | 20 scenarios | CRUD operations, validation, errors |
| Authorization | 8 scenarios | Data isolation, security, token validation |
| Error Handling | 12 scenarios | Edge cases, malformed requests, server errors |
| Performance | 6 scenarios | Load testing, concurrent operations |
| Integration | 4 scenarios | End-to-end workflows, multi-user scenarios |
| Concurrent Updates | 4 scenarios | Race conditions, data loss, unpredictable behavior |

## Expected Outcomes

### Positive Test Cases
- All happy path scenarios should pass
- Users can successfully register, login, and manage todos
- Data isolation works correctly
- Performance meets acceptable thresholds

### Negative Test Cases
- Error handling provides meaningful responses
- Invalid requests are rejected appropriately
- Security measures prevent unauthorized access
- Race conditions demonstrate expected vulnerabilities

### Risk Areas
- **High Risk:** Concurrent update race conditions (intentional)
- **Medium Risk:** Authentication token handling
- **Low Risk:** Basic CRUD operations

## Maintenance

### Adding New Scenarios
1. Identify the appropriate feature file
2. Add new scenario with Given/When/Then structure
3. Update step definitions if needed
4. Update this README with new scenario count

### Updating Existing Scenarios
1. Modify the scenario in the appropriate feature file
2. Update step definitions if behavior changes
3. Test the updated scenario
4. Update documentation if needed

## Notes

- The concurrent update scenarios are designed to demonstrate race conditions and may fail intentionally
- Performance scenarios may require specific test environment setup
- Integration scenarios test complete user workflows and may take longer to execute
- All scenarios include proper cleanup and isolation between tests