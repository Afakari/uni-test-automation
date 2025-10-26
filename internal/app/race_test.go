package app

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// The helper functions performRequest and Init are accessible from handlers_test.go
// because they are in the same package (app).

// setupRaceRouter is not strictly needed since SetupRouter already exists,
// but we keep the helper function 'performRequest' (from handlers_test.go)
// which is used by this test.

// TestConcurrentUpdateRace is the key test to showcase the data race.
// It will only fail/warn when run with the Go race detector flag (-race).
func TestConcurrentUpdateRace(t *testing.T) {
	// Reset global maps
	Users = sync.Map{}
	Todos = sync.Map{}

	// Ensure the secret is initialized for JWT operations
	testSecret := "testsecret_race"
	os.Setenv("JWT_SECRET", testSecret)
	Init(testSecret)
	gin.SetMode(gin.TestMode)

	r := SetupRouter()

	// --- Setup user & todo ---
	// Using the helper from handlers_test.go
	performRequest(r, "POST", "/register", Credentials{"racer", "pass"}, "")
	w := performRequest(r, "POST", "/login", Credentials{"racer", "pass"}, "")
	var loginResp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &loginResp)
	token := loginResp["token"]

	// Create todo
	w = performRequest(r, "POST", "/todos", map[string]string{"title": "base task"}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("Create failed: %s", w.Body.String())
	}
	var created Todo
	_ = json.Unmarshal(w.Body.Bytes(), &created)

	// --- Run two concurrent updates ---
	const loops = 500
	wg := sync.WaitGroup{}
	wg.Add(2)

	updateFunc := func(titlePrefix string) {
		defer wg.Done()
		for i := 0; i < loops; i++ {
			// Ensure unique values for each update
			newTitle := titlePrefix + " " + time.Now().Format("15:04:05.000000")
			body := UpdateTodoRequest{Title: &newTitle}
			// This call hits the UpdateTodoHandler which has the race (found.Title = *req.Title)
			performRequest(r, "PUT", "/todos/"+created.ID, body, token)
		}
	}

	t.Logf("Starting %d concurrent updates to trigger the data race on UpdateTodoHandler...", loops)
	go updateFunc("VERSION A")
	go updateFunc("VERSION B")
	wg.Wait()
	t.Log("Concurrent updates finished.")

	// --- Get final state ---
	w = performRequest(r, "GET", "/todos/"+created.ID, nil, token)
	var final Todo
	_ = json.Unmarshal(w.Body.Bytes(), &final)
	t.Logf("Final todo title: %s", final.Title)
}
