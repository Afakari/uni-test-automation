package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"todoapp/internal/app"
)

var concurrentWG sync.WaitGroup
var stateMutex sync.Mutex

var tf *todoFeature

// state
type todoFeature struct {
	server        *httptest.Server
	client        *http.Client
	baseURL       string
	token         string // Used primarily by concurrent_update feature
	todoID        string
	originalTitle string
	errs          []error
	lastResponse  *http.Response
	lastErrorMsg  string
	success       int

	// Fields for multi-user support (Authorization feature)
	userTokens    map[string]string // Key: username, Value: JWT
	userTodoIDs   map[string]string // Key: username, Value: latest TodoID
	currentUser   string            // Tracks which user is active for the next request
	todoIDByTitle map[string]string
}

// 🎯 REVISED makeRequest: Multi-User Token Logic
func (tf *todoFeature) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	var buf io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		buf = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, tf.baseURL+path, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// --- Multi-User Token Logic ---
	var tokenToUse string

	// 1. Check if a currentUser is set and has a token (for new Auth feature)
	if tf.currentUser != "" {
		if t, ok := tf.userTokens[tf.currentUser]; ok {
			tokenToUse = t
		}
	}

	// 2. Fallback to the generic tf.token (for existing concurrent_update logic)
	if tokenToUse == "" {
		tokenToUse = tf.token
	}

	// 3. Set the Authorization header if any token was found
	if tokenToUse != "" {
		req.Header.Set("Authorization", "Bearer "+tokenToUse)
	}
	// ------------------------------

	return tf.client.Do(req)
}

func (tf *todoFeature) recordAppendAttempt(charToAppend string) {
	// Logic remains unchanged, it correctly uses tf.makeRequest
	// which will now correctly apply tf.token (as currentUser is likely not set
	// or tf.token was explicitly set during login for the concurrent test).

	// 1. READ
	getResp, err := tf.makeRequest("GET", "/todos/"+tf.todoID, nil)
	if err != nil {
		stateMutex.Lock()
		tf.errs = append(tf.errs, fmt.Errorf("GET request error: %w", err))
		stateMutex.Unlock()
		return
	}

	var currentTodo app.Todo
	if err := json.NewDecoder(getResp.Body).Decode(&currentTodo); err != nil {
		_ = getResp.Body.Close()
		stateMutex.Lock()
		tf.errs = append(tf.errs, fmt.Errorf("GET decode error: %w", err))
		stateMutex.Unlock()
		return
	}
	// IMPORTANT: Close body after read
	_ = getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		stateMutex.Lock()
		tf.errs = append(tf.errs, fmt.Errorf("GET failed with status %d", getResp.StatusCode))
		stateMutex.Unlock()
		return
	}

	newTitle := currentTodo.Title + charToAppend
	req := app.UpdateTodoRequest{Title: &newTitle}

	putResp, err := tf.makeRequest("PUT", "/todos/"+tf.todoID, req)
	if err != nil {
		stateMutex.Lock()
		tf.errs = append(tf.errs, fmt.Errorf("PUT request error: %w", err))
		stateMutex.Unlock()
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(putResp.Body)

	stateMutex.Lock()
	if putResp.StatusCode == http.StatusOK {
		tf.success++
	} else {
		tf.errs = append(tf.errs, fmt.Errorf("PUT update failed with status %d", putResp.StatusCode))
	}
	stateMutex.Unlock()
}

// setupServer logic is good, keeping it as a method on tf
func (tf *todoFeature) setupServer(ctx context.Context) (context.Context, error) {
	router := app.SetupRouter()

	tf.server = httptest.NewServer(router)
	tf.baseURL = tf.server.URL
	tf.client = tf.server.Client()
	tf.errs = nil
	tf.success = 0
	tf.lastResponse = nil
	tf.lastErrorMsg = ""
	tf.userTokens = make(map[string]string)
	tf.userTodoIDs = make(map[string]string)
	tf.currentUser = ""
	return ctx, nil
}

// closeSever logic is good, keeping it as a method on tf
func (tf *todoFeature) closeSever(ctx context.Context) (context.Context, error) {
	if tf.server != nil {
		tf.server.Close()
	}
	return ctx, nil
}
