# BDD Testing Best Practices for Go Todo Application

## Overview
This document outlines the best practices for Behavior-Driven Development (BDD) testing using Gherkin/Cucumber with godog in our Go todo application.

## 1. Feature File Organization

### Directory Structure
```
internal/features/
├── authentication.feature
├── authorization.feature
├── todo_management.feature
├── concurrent_operations.feature
├── integration.feature
├── steps/
│   ├── authentication_steps.go
│   ├── authorization_steps.go
│   ├── todo_management_steps.go
│   ├── concurrent_operations_steps.go
│   ├── integration_steps.go
│   └── common_steps.go
└── support/
    ├── test_context.go
    ├── test_helpers.go
    └── test_config.go
```

### Feature File Best Practices

#### 1.1 Feature Structure
```gherkin
Feature: [Feature Name]
  In order to [business value]
  As a [user role]
  I want to [capability]

  Background:
    Given [common setup]

  Scenario: [Scenario name]
    Given [initial context]
    When [action]
    Then [expected outcome]
    And [additional verification]
```

#### 1.2 Scenario Naming
- Use descriptive, business-focused names
- Avoid technical jargon
- Focus on user behavior, not implementation
- Examples:
  - ✅ "User can create a new todo item"
  - ❌ "POST /todos endpoint returns 201"

#### 1.3 Background vs Given
- **Background**: Common setup for all scenarios in a feature
- **Given**: Specific setup for individual scenarios
- Use Background for authentication, common data setup
- Use Given for scenario-specific preconditions

## 2. Step Definition Best Practices

### 2.1 Step Organization
```go
// Group related steps together
func InitializeTodoManagementScenario(ctx *godog.ScenarioContext) {
    // Setup and teardown
    ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
        return setupTestContext(ctx)
    })
    
    ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
        return cleanupTestContext(ctx)
    })
    
    // Given steps
    ctx.Step(`^a user named "([^"]*)" is logged in$`, userIsLoggedIn)
    ctx.Step(`^the user has no todos$`, userHasNoTodos)
    
    // When steps
    ctx.Step(`^the user creates a todo with title "([^"]*)"$`, userCreatesTodo)
    ctx.Step(`^the user updates the todo title to "([^"]*)"$`, userUpdatesTodo)
    
    // Then steps
    ctx.Step(`^the user should see (\d+) todos$`, userShouldSeeTodos)
    ctx.Step(`^the todo should have title "([^"]*)"$`, todoShouldHaveTitle)
}
```

### 2.2 Step Implementation Guidelines

#### Use Context for State Management
```go
type TestContext struct {
    Server      *httptest.Server
    Client      *http.Client
    UserTokens  map[string]string
    TodoIDs     map[string]string
    LastResponse *http.Response
    LastError   string
}

func setupTestContext(ctx context.Context) (context.Context, error) {
    testCtx := &TestContext{
        UserTokens: make(map[string]string),
        TodoIDs:    make(map[string]string),
    }
    return context.WithValue(ctx, "testContext", testCtx), nil
}
```

#### Implement Reusable Helper Functions
```go
func (tc *TestContext) makeAuthenticatedRequest(method, path string, body interface{}, username string) (*http.Response, error) {
    token, exists := tc.UserTokens[username]
    if !exists {
        return nil, fmt.Errorf("user %s is not authenticated", username)
    }
    
    return tc.makeRequest(method, path, body, token)
}

func (tc *TestContext) makeRequest(method, path string, body interface{}, token string) (*http.Response, error) {
    // Implementation details...
}
```

### 2.3 Error Handling
```go
func userCreatesTodo(ctx context.Context, title string) (context.Context, error) {
    tc := getTestContext(ctx)
    
    todo := map[string]string{"title": title}
    resp, err := tc.makeAuthenticatedRequest("POST", "/todos", todo, tc.CurrentUser)
    if err != nil {
        return ctx, fmt.Errorf("failed to create todo: %w", err)
    }
    
    tc.LastResponse = resp
    return ctx, nil
}
```

## 3. Test Data Management

### 3.1 Test Data Strategy
- Use realistic but anonymized data
- Create data factories for consistent test data
- Clean up test data between scenarios
- Use unique identifiers to avoid conflicts

### 3.2 Data Factory Pattern
```go
type TestDataFactory struct {
    UserCounter int
    TodoCounter int
}

func (f *TestDataFactory) CreateUser() (string, string) {
    f.UserCounter++
    return fmt.Sprintf("user%d", f.UserCounter), "password123"
}

func (f *TestDataFactory) CreateTodo() string {
    f.TodoCounter++
    return fmt.Sprintf("Test Todo %d", f.TodoCounter)
}
```

## 4. Test Organization and Configuration

### 4.1 Test Suite Structure
```go
func TestFeatureSuite(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeFeatureScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"../feature.feature"},
            TestingT: t,
        },
    }
    
    if suite.Run() != 0 {
        t.Fatal("feature tests failed")
    }
}
```

### 4.2 Test Configuration
```go
type TestConfig struct {
    BaseURL     string
    JWTSecret   string
    Timeout     time.Duration
    RetryCount  int
}

func LoadTestConfig() *TestConfig {
    return &TestConfig{
        BaseURL:    getEnv("TEST_BASE_URL", "http://localhost:8080"),
        JWTSecret:  getEnv("JWT_SECRET", "test-secret"),
        Timeout:    30 * time.Second,
        RetryCount: 3,
    }
}
```

## 5. Assertion Best Practices

### 5.1 Clear and Descriptive Assertions
```go
func todoShouldHaveTitle(ctx context.Context, expectedTitle string) (context.Context, error) {
    tc := getTestContext(ctx)
    
    if tc.LastResponse == nil {
        return ctx, fmt.Errorf("no response available for assertion")
    }
    
    var todo Todo
    if err := json.NewDecoder(tc.LastResponse.Body).Decode(&todo); err != nil {
        return ctx, fmt.Errorf("failed to decode todo response: %w", err)
    }
    
    if todo.Title != expectedTitle {
        return ctx, fmt.Errorf("expected todo title '%s', got '%s'", expectedTitle, todo.Title)
    }
    
    return ctx, nil
}
```

### 5.2 Comprehensive Response Validation
```go
func responseShouldHaveStatus(ctx context.Context, expectedStatus int) (context.Context, error) {
    tc := getTestContext(ctx)
    
    if tc.LastResponse.StatusCode != expectedStatus {
        return ctx, fmt.Errorf("expected status %d, got %d", expectedStatus, tc.LastResponse.StatusCode)
    }
    
    return ctx, nil
}
```

## 6. Performance and Reliability

### 6.1 Test Isolation
- Each scenario should be independent
- Clean up state between scenarios
- Use unique test data to avoid conflicts
- Reset application state when needed

### 6.2 Timeout and Retry Logic
```go
func (tc *TestContext) makeRequestWithRetry(method, path string, body interface{}, token string) (*http.Response, error) {
    for i := 0; i < tc.Config.RetryCount; i++ {
        resp, err := tc.makeRequest(method, path, body, token)
        if err == nil && resp.StatusCode < 500 {
            return resp, nil
        }
        
        if i < tc.Config.RetryCount-1 {
            time.Sleep(time.Duration(i+1) * time.Second)
        }
    }
    
    return nil, fmt.Errorf("request failed after %d retries", tc.Config.RetryCount)
}
```

## 7. Documentation and Maintenance

### 7.1 Living Documentation
- Feature files serve as living documentation
- Keep scenarios up-to-date with business requirements
- Use business language, not technical terms
- Include examples and edge cases

### 7.2 Test Maintenance
- Regular review of test scenarios
- Remove obsolete tests
- Refactor common patterns into reusable steps
- Keep step definitions focused and single-purpose

## 8. Common Anti-Patterns to Avoid

### 8.1 Technical Steps
```gherkin
# ❌ Bad - too technical
When I send a POST request to "/todos" with JSON body {"title": "Test"}

# ✅ Good - business focused
When I create a todo with title "Test"
```

### 8.2 Overly Complex Scenarios
```gherkin
# ❌ Bad - too many steps
Scenario: Complex user journey
  Given a user is logged in
  And the user has 5 todos
  And the user has 3 completed todos
  And the user has 2 pending todos
  When the user creates a new todo
  And the user updates the first todo
  And the user deletes the second todo
  And the user marks the third todo as complete
  Then the user should see 4 todos
  And the user should see 4 completed todos
  And the user should see 0 pending todos

# ✅ Good - focused scenarios
Scenario: User can create a todo
  Given a user is logged in
  When the user creates a todo with title "New task"
  Then the user should see 1 todo with title "New task"

Scenario: User can update a todo
  Given a user is logged in
  And the user has a todo with title "Original task"
  When the user updates the todo title to "Updated task"
  Then the user should see 1 todo with title "Updated task"
```

### 8.3 Tight Coupling
- Avoid hardcoded values in step definitions
- Use configuration for environment-specific values
- Make steps reusable across different features
- Keep business logic in the application, not in tests

## 9. Running and Reporting

### 9.1 Test Execution
```bash
# Run all BDD tests
go test ./internal/features/...

# Run specific feature
go test -run TestAuthentication ./internal/features/steps/

# Run with verbose output
go test -v ./internal/features/...

# Run with custom format
go test -godog.format=progress ./internal/features/...
```

### 9.2 Continuous Integration
- Include BDD tests in CI pipeline
- Run tests on multiple environments
- Generate test reports
- Fail builds on test failures

## 10. Benefits of Following These Practices

1. **Maintainability**: Well-organized tests are easier to maintain
2. **Readability**: Clear scenarios serve as documentation
3. **Reliability**: Isolated tests are more reliable
4. **Collaboration**: Business stakeholders can understand and contribute
5. **Coverage**: Comprehensive scenarios catch more bugs
6. **Confidence**: Well-tested features give confidence in deployments

This guide should help you maintain high-quality BDD tests that serve as both documentation and validation for your todo application.
