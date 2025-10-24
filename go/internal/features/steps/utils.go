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

// state
type todoFeature struct {
	server        *httptest.Server
	client        *http.Client
	baseURL       string
	token         string
	todoID        string
	originalTitle string
	errs          []error
	success       int
}

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
	if tf.token != "" {
		req.Header.Set("Authorization", "Bearer "+tf.token)
	}

	return tf.client.Do(req)
}

func (tf *todoFeature) recordAppendAttempt(charToAppend string) {
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
