package steps

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

// TestContext holds the state for ALL BDD tests
type TestContext struct {
	Server        *httptest.Server
	Client        *http.Client
	BaseURL       string
	Config        *TestConfig
	UserTokens    map[string]string
	UserTodoIDs   map[string]string
	CurrentUser   string
	LastResponse  *http.Response
	LastError     string
	TodoIDByTitle map[string]string
	mutex         sync.RWMutex

	// --- Fields for concurrent_update_test ---
	CurrentTodoID string
	OriginalTitle string
	success       int
	errs          []error
	concMutex     sync.Mutex
	// -----------------------------------------
}

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
		errs:          make([]error, 0),
	}
}

func LoadTestConfig() *TestConfig {
	return &TestConfig{
		JWTSecret:     getEnv("JWT_SECRET", "test-secret"),
		BaseURL:       getEnv("TEST_BASE_URL", ""),
		Timeout:       getDurationEnv("TEST_TIMEOUT", 30*time.Second),
		RetryCount:    getIntEnv("TEST_RETRY_COUNT", 3),
		LogLevel:      getEnv("TEST_LOG_LEVEL", "info"),
		TestDataDir:   getEnv("TEST_DATA_DIR", "./testdata"),
		CleanupAfter:  getBoolEnv("TEST_CLEANUP_AFTER", true),
		ParallelTests: getBoolEnv("TEST_PARALLEL", false),
	}
}

func (tc *TestContext) SetupServer() error {
	app.Init(tc.Config.JWTSecret)
	app.Users = sync.Map{}
	app.Todos = sync.Map{}

	router := app.SetupRouter()
	tc.Server = httptest.NewServer(router)
	tc.BaseURL = tc.Server.URL
	tc.Client = &http.Client{Timeout: tc.Config.Timeout}
	return nil
}

func (tc *TestContext) CloseServer() {
	if tc.Server != nil {
		tc.Server.Close()
	}
}

func (tc *TestContext) MakeRequest(method, path string, body interface{}) (*http.Response, error) {
	var buf io.Reader
	if body != nil {
		if strBody, ok := body.(string); ok {
			buf = bytes.NewReader([]byte(strBody))
		} else {
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			buf = bytes.NewReader(jsonBody)
		}
	}

	req, err := http.NewRequestWithContext(context.Background(), method, tc.BaseURL+path, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if tc.CurrentUser != "" {
		if token, exists := tc.UserTokens[tc.CurrentUser]; exists {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	return tc.Client.Do(req)
}

func (tc *TestContext) SetCurrentUser(username string) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.CurrentUser = username
}

func (tc *TestContext) StoreUserToken(username, token string) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.UserTokens[username] = token
}

func (tc *TestContext) GetUserToken(username string) (string, bool) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	token, exists := tc.UserTokens[username]
	return token, exists
}

func (tc *TestContext) StoreTodoID(username, todoID string) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.UserTodoIDs[username] = todoID
}

func (tc *TestContext) GetTodoID(username string) (string, bool) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	todoID, exists := tc.UserTodoIDs[username]
	return todoID, exists
}

func (tc *TestContext) StoreTodoByTitle(title, todoID string) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.TodoIDByTitle[title] = todoID
}

func (tc *TestContext) GetTodoIDByTitle(title string) (string, bool) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	todoID, exists := tc.TodoIDByTitle[title]
	return todoID, exists
}

func (tc *TestContext) SetLastResponse(resp *http.Response) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.LastResponse = resp
}

func (tc *TestContext) GetLastResponse() *http.Response {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	return tc.LastResponse
}

type contextKey string

const testContextKey contextKey = "testContext"

func GetTestContextFromContext(ctx context.Context) *TestContext {
	if tc, ok := ctx.Value(testContextKey).(*TestContext); ok {
		return tc
	}
	return nil
}

func SetTestContextInContext(ctx context.Context, tc *TestContext) context.Context {
	return context.WithValue(ctx, testContextKey, tc)
}

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
