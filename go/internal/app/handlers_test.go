package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

// init runs before any test in this package
func init() {
	// Set the testing mode for gin
	gin.SetMode(gin.TestMode)

	// Set the JWT secret for the application logic
	testSecret := "testsecret_standard"
	os.Setenv("JWT_SECRET", testSecret)
	Init(testSecret)
}

// performRequest is a helper to execute HTTP requests on the router
func performRequest(r *gin.Engine, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// TestFullCRUDFlow runs all standard non-concurrent operations.
// This test passes whether or not a race condition exists.
func TestFullCRUDFlow(t *testing.T) {
	// Reset global maps for isolated test run
	Users = sync.Map{}
	Todos = sync.Map{}

	// SetupRouter is in scope and provides the configured router
	r := SetupRouter()

	// 1) Register
	regBody := Credentials{Username: "alice", Password: "pass"}
	w := performRequest(r, "POST", "/register", regBody, "")
	if w.Code != http.StatusCreated {
		t.Fatalf("Register failed: %d body=%s", w.Code, w.Body.String())
	}

	// 2) Login
	w = performRequest(r, "POST", "/login", regBody, "")
	if w.Code != http.StatusOK {
		t.Fatalf("Login failed: %d body=%s", w.Code, w.Body.String())
	}
	var loginResp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &loginResp)
	token := loginResp["token"]
	if token == "" {
		t.Fatal("No token returned")
	}

	// 3) Create Todo
	createReq := map[string]string{"title": "first task"}
	w = performRequest(r, "POST", "/todos", createReq, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("Create todo failed: %d body=%s", w.Code, w.Body.String())
	}
	var created Todo
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	if created.ID == "" {
		t.Fatal("Created todo missing ID")
	}

	// 4) Get Todos
	w = performRequest(r, "GET", "/todos", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("Get todos failed: %d body=%s", w.Code, w.Body.String())
	}
	var list []Todo
	_ = json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 {
		t.Fatalf("Expected 1 todo got %d", len(list))
	}

	// 5) Update Todo (sequential)
	title := "updated title"
	upd := UpdateTodoRequest{Title: &title}
	w = performRequest(r, "PUT", "/todos/"+created.ID, upd, token)
	if w.Code != http.StatusOK {
		t.Fatalf("Update todo failed: %d body=%s", w.Code, w.Body.String())
	}
	var updated Todo
	_ = json.Unmarshal(w.Body.Bytes(), &updated)
	if updated.Title != title {
		t.Fatalf("Expected title %q got %q", title, updated.Title)
	}

	// 6) Delete Todo
	w = performRequest(r, "DELETE", "/todos/"+created.ID, nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("Delete todo failed: %d body=%s", w.Code, w.Body.String())
	}

	// 7) Ensure gone
	w = performRequest(r, "GET", "/todos", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("Get todos after delete failed: %d body=%s", w.Code, w.Body.String())
	}
	_ = json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 0 {
		t.Fatalf("Expected 0 todos got %d", len(list))
	}
}
