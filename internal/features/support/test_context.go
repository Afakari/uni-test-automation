package support

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"time"

	"todoapp/internal/app"
)

// TestContext holds the state for BDD tests
type TestContext struct {
	Server        *httptest.Server
	Client        *http.Client
	BaseURL       string
	Config        *TestConfig
	UserTokens    map[string]string // username -> JWT token
	UserTodoIDs   map[string]string // username -> latest todo ID
	CurrentUser   string            // tracks which user is active
	LastResponse  *http.Response
	LastError     string
	TodoIDByTitle map[string]string // title -> todo ID for lookups
	mutex         sync.RWMutex
}

// TestConfig holds test configuration
type TestConfig struct {
	JWTSecret     string
	BaseURL       string
	Timeout       time.Duration
	RetryCount    int
	LogLevel      string
	TestDataDir   string
	CleanupAfter  bool
	ParallelTests bool
}

// NewTestContext creates a new test context
func NewTestContext() *TestContext {
	config := LoadTestConfig()
	return &TestContext{
		UserTokens:    make(map[string]string),
		UserTodoIDs:   make(map[string]string),
		TodoIDByTitle: make(map[string]string),
		Config:        config,
	}
}

// LoadTestConfig loads test configuration from environment variables with defaults
func LoadTestConfig() *TestConfig {
	return &TestConfig{
		JWTSecret:     getEnv("JWT_SECRET", "test-secret"),
		BaseURL:       getEnv("TEST_BASE_URL", ""), // Will be set dynamically for test server
		Timeout:       getDurationEnv("TEST_TIMEOUT", 30*time.Second),
		RetryCount:    getIntEnv("TEST_RETRY_COUNT", 3),
		LogLevel:      getEnv("TEST_LOG_LEVEL", "info"),
		TestDataDir:   getEnv("TEST_DATA_DIR", "./testdata"),
		CleanupAfter:  getBoolEnv("TEST_CLEANUP_AFTER", true),
		ParallelTests: getBoolEnv("TEST_PARALLEL", false),
	}
}

// SetupServer initializes the test server
func (tc *TestContext) SetupServer() error {
	router := app.SetupRouter()
	tc.Server = httptest.NewServer(router)
	tc.BaseURL = tc.Server.URL
	tc.Client = &http.Client{Timeout: tc.Config.Timeout}
	return nil
}

// CloseServer cleans up the test server
func (tc *TestContext) CloseServer() {
	if tc.Server != nil {
		tc.Server.Close()
	}
}

// MakeRequest creates and executes an HTTP request
func (tc *TestContext) MakeRequest(method, path string, body interface{}) (*http.Response, error) {
	var buf io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		buf = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, tc.BaseURL+path, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Add authentication if user is set
	if tc.CurrentUser != "" {
		if token, exists := tc.UserTokens[tc.CurrentUser]; exists {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	return tc.Client.Do(req)
}

// MakeAuthenticatedRequest creates a request with authentication for a specific user
func (tc *TestContext) MakeAuthenticatedRequest(method, path string, body interface{}, username string) (*http.Response, error) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	token, exists := tc.UserTokens[username]
	if !exists {
		return nil, fmt.Errorf("user %s is not authenticated", username)
	}

	var buf io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		buf = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, tc.BaseURL+path, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	return tc.Client.Do(req)
}

// SetCurrentUser sets the active user for subsequent requests
func (tc *TestContext) SetCurrentUser(username string) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.CurrentUser = username
}

// StoreUserToken stores a JWT token for a user
func (tc *TestContext) StoreUserToken(username, token string) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.UserTokens[username] = token
}

// GetUserToken retrieves a JWT token for a user
func (tc *TestContext) GetUserToken(username string) (string, bool) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	token, exists := tc.UserTokens[username]
	return token, exists
}

// StoreTodoID stores a todo ID for a user
func (tc *TestContext) StoreTodoID(username, todoID string) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.UserTodoIDs[username] = todoID
}

// GetTodoID retrieves the latest todo ID for a user
func (tc *TestContext) GetTodoID(username string) (string, bool) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	todoID, exists := tc.UserTodoIDs[username]
	return todoID, exists
}

// StoreTodoByTitle stores a mapping from title to todo ID
func (tc *TestContext) StoreTodoByTitle(title, todoID string) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.TodoIDByTitle[title] = todoID
}

// GetTodoIDByTitle retrieves a todo ID by title
func (tc *TestContext) GetTodoIDByTitle(title string) (string, bool) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	todoID, exists := tc.TodoIDByTitle[title]
	return todoID, exists
}

// SetLastResponse stores the last HTTP response
func (tc *TestContext) SetLastResponse(resp *http.Response) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.LastResponse = resp
}

// GetLastResponse retrieves the last HTTP response
func (tc *TestContext) GetLastResponse() *http.Response {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	return tc.LastResponse
}

// SetLastError stores the last error message
func (tc *TestContext) SetLastError(err string) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.LastError = err
}

// GetLastError retrieves the last error message
func (tc *TestContext) GetLastError() string {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	return tc.LastError
}

// ClearState resets the test context state
func (tc *TestContext) ClearState() {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.UserTokens = make(map[string]string)
	tc.UserTodoIDs = make(map[string]string)
	tc.TodoIDByTitle = make(map[string]string)
	tc.CurrentUser = ""
	tc.LastResponse = nil
	tc.LastError = ""
}

// Context key for storing test context
type contextKey string

const testContextKey contextKey = "testContext"

// GetTestContextFromContext retrieves the test context from a context.Context
func GetTestContextFromContext(ctx context.Context) *TestContext {
	if tc, ok := ctx.Value(testContextKey).(*TestContext); ok {
		return tc
	}
	return nil
}

// SetTestContextInContext stores the test context in a context.Context
func SetTestContextInContext(ctx context.Context, tc *TestContext) context.Context {
	return context.WithValue(ctx, testContextKey, tc)
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
