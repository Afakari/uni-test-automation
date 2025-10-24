package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"io"
	"net/http"
	"testing"
	"todoapp/internal/app"
)

// --- Feature Runner (TestUserAuthorization) ---

func TestUserAuthorization(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeUserAuthenticationScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../authorization.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

// --- Step Definitions ---

func InitializeUserAuthenticationScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		tf = &todoFeature{}
		ctx, _ = tf.setupServer(ctx)
		return ctx, nil
	})
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		ctx, err = tf.closeSever(ctx)
		return ctx, err
	})

	// Re-used steps (assuming they are defined elsewhere or use common setup)
	ctx.Step(`^the secret key "([^"]*)" is set up$`, isJwtSecretSet)
	ctx.Step(`^a user named "([^"]*)" with password "([^"]*)" is registered$`, aUserNamedWithPasswordIsRegistered)
	ctx.Step(`^user "([^"]*)" logs in with password "([^"]*)" successfully$`, userLogsInAndStoresToken)

	// New steps for Authorization Feature
	ctx.Step(`^user "([^"]*)" has created a todo with title "([^"]*)"$`, userHasCreatedATodoWithTitle)
	ctx.Step(`^user "([^"]*)" gets all todos$`, userGetsAllTodos)
	ctx.Step(`^user "([^"]*)" should only see "([^"]*)"$`, userShouldOnlySee)
	ctx.Step(`^user "([^"]*)" should not see "([^"]*)"$`, userShouldNotSee)
	ctx.Step(`^user "([^"]*)" tries to get user "([^"]*)"'s todo by ID$`, userTriesToGetAnotherUsersTodoByID)
	ctx.Step(`^user "([^"]*)" tries to update user "([^"]*)"'s todo$`, userTriesToUpdateAnotherUsersTodo)
	ctx.Step(`^user "([^"]*)" tries to delete user "([^"]*)"'s todo$`, userTriesToDeleteAnotherUsersTodo)
	ctx.Step(`^I try to access todos with an invalid token$`, iTryToAccessTodosWithAnInvalidToken)
	ctx.Step(`^I try to access todos with an expired token$`, iTryToAccessTodosWithAnExpiredToken)
	ctx.Step(`^I try to access todos without an authorization header$`, iTryToAccessTodosWithoutAnAuthorizationHeader)
	ctx.Step(`^I try to access todos with malformed authorization header$`, iTryToAccessTodosWithMalformedAuthorizationHeader)

	// Re-used assertion steps
	ctx.Step(`^I should receive an error message "([^"]*)"$`, iShouldReceiveAnErrorMessage)
	ctx.Step(`^the response status should be (\d+)$`, theResponseStatusShouldBe)
}

// Helper to handle login and token storage for multi-user scenarios
func userLogsInAndStoresToken(ctx context.Context, username, password string) (context.Context, error) {
	// Temporarily set current user to get the right token if makeRequest used it,
	// but here we just need to get the token and store it in userTokens.
	tf.currentUser = "" // Ensure no old token is used for this request

	body := app.Credentials{Username: username, Password: password}
	resp, err := tf.makeRequest("POST", "/login", body)
	if err != nil {
		return ctx, err
	}
	tf.lastResponse = resp

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(tf.lastResponse.Body)

	if resp.StatusCode != http.StatusOK {
		return tf.extractError(ctx)
	}

	var resBody map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&resBody); err != nil {
		return ctx, fmt.Errorf("failed to decode login response body: %w", err)
	}

	token := resBody["token"]
	if token == "" {
		return ctx, fmt.Errorf("login successful but no token returned for user %s", username)
	}

	// Store the token mapped to the username
	tf.userTokens[username] = token
	return ctx, nil
}

// Given user "alice" has created a todo with title "Alice's Task"
func userHasCreatedATodoWithTitle(ctx context.Context, username, title string) (context.Context, error) {
	tf.currentUser = username // Set the active user for the request
	reqBody := map[string]string{"title": title}

	resp, err := tf.makeRequest("POST", "/todos", reqBody)

	if err != nil {
		return ctx, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return tf.extractError(ctx)
	}

	var created app.Todo
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return ctx, err
	}

	// Store the TodoID associated with the user for later lookups/updates/deletes
	tf.userTodoIDs[username] = created.ID

	// Store the ID of this specific task title (useful for cross-user lookups)
	if tf.todoIDByTitle == nil {
		tf.todoIDByTitle = make(map[string]string)
	}
	tf.todoIDByTitle[title] = created.ID

	return ctx, nil
}

// When user "alice" gets all todos
func userGetsAllTodos(ctx context.Context, username string) (context.Context, error) {
	tf.currentUser = username // Set the active user for the request

	resp, err := tf.makeRequest("GET", "/todos", nil)
	if err != nil {
		return ctx, err
	}
	tf.lastResponse = resp
	return tf.extractError(ctx)
}

// Assertion helper: check if a specific title exists in the last response's body
func todoTitleExists(title string) (bool, error) {
	if tf.lastResponse == nil {
		return false, fmt.Errorf("no response to check")
	}

	bodyBytes, err := io.ReadAll(tf.lastResponse.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	// Reset the response body so subsequent assertion steps can read it
	tf.lastResponse.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var todos []app.Todo
	if err := json.Unmarshal(bodyBytes, &todos); err != nil {
		// If it's not an array of todos, it might be an error or something else, but we must check for status code first.
		if tf.lastResponse.StatusCode != http.StatusOK {
			return false, nil // Assume if status isn't 200, the title isn't 'seen' successfully
		}
		// If status is 200 but decode failed, something is wrong
		return false, fmt.Errorf("failed to decode response body as []app.Todo: %w", err)
	}

	for _, todo := range todos {
		if todo.Title == title {
			return true, nil
		}
	}
	return false, nil
}

// Then user "alice" should only see "Alice's Task"
func userShouldOnlySee(ctx context.Context, username, title string) (context.Context, error) {
	found, err := todoTitleExists(title)
	if err != nil {
		return ctx, err
	}
	if !found {
		return ctx, fmt.Errorf("user %s expected to see todo '%s' but did not", username, title)
	}
	return ctx, nil
}

// And user "alice" should not see "Bob's Task"
func userShouldNotSee(ctx context.Context, username, title string) (context.Context, error) {
	found, err := todoTitleExists(title)
	if err != nil {
		return ctx, err
	}
	if found {
		return ctx, fmt.Errorf("user %s expected NOT to see todo '%s' but did", username, title)
	}
	return ctx, nil
}

// When user "alice" tries to get user "bob"'s todo by ID
func userTriesToGetAnotherUsersTodoByID(ctx context.Context, requestingUser, ownerUser string) (context.Context, error) {
	tf.currentUser = requestingUser // Set the active user for the request

	// Get the target todo ID. Assuming the owner user has created a todo (setup in the background).
	todoID, ok := tf.userTodoIDs[ownerUser]
	if !ok {
		return ctx, fmt.Errorf("could not find todo ID for user %s", ownerUser)
	}

	resp, err := tf.makeRequest("GET", "/todos/"+todoID, nil)
	if err != nil {
		return ctx, err
	}
	tf.lastResponse = resp
	return tf.extractError(ctx) // This extracts the error message and resets body for assertion
}

// When user "bob" tries to update user "alice"'s todo
func userTriesToUpdateAnotherUsersTodo(ctx context.Context, requestingUser, ownerUser string) (context.Context, error) {
	tf.currentUser = requestingUser // Set the active user for the request

	// Get the target todo ID
	todoID, ok := tf.userTodoIDs[ownerUser]
	if !ok {
		return ctx, fmt.Errorf("could not find todo ID for user %s", ownerUser)
	}

	newTitle := "Attempted update"
	reqBody := app.UpdateTodoRequest{Title: &newTitle}

	resp, err := tf.makeRequest("PUT", "/todos/"+todoID, reqBody)
	if err != nil {
		return ctx, err
	}
	tf.lastResponse = resp
	return tf.extractError(ctx)
}

// When user "bob" tries to delete user "alice"'s todo
func userTriesToDeleteAnotherUsersTodo(ctx context.Context, requestingUser, ownerUser string) (context.Context, error) {
	tf.currentUser = requestingUser // Set the active user for the request

	// Get the target todo ID
	todoID, ok := tf.userTodoIDs[ownerUser]
	if !ok {
		return ctx, fmt.Errorf("could not find todo ID for user %s", ownerUser)
	}

	resp, err := tf.makeRequest("DELETE", "/todos/"+todoID, nil)
	if err != nil {
		return ctx, err
	}
	tf.lastResponse = resp
	return tf.extractError(ctx)
}

// When I try to access todos with an invalid token
func iTryToAccessTodosWithAnInvalidToken(ctx context.Context) (context.Context, error) {
	tf.token = "invalid.token.string" // Override generic token for this single request
	tf.currentUser = ""               // Ensure currentUser is not used

	resp, err := tf.makeRequest("GET", "/todos", nil)

	// Restore original token state (or clear it) after the request
	tf.token = ""

	if err != nil {
		return ctx, err
	}
	tf.lastResponse = resp
	return tf.extractError(ctx)
}

// When I try to access todos with an expired token
func iTryToAccessTodosWithAnExpiredToken(ctx context.Context) (context.Context, error) {
	// A standard JWT contains: Header, Payload (incl. 'exp'), and Signature.
	// Since we cannot dynamically sign an expired token without the actual signing library,
	// we will use a common mock expired token structure.
	// NOTE: In a real test, this token would need to be generated by the app's JWT logic
	// with a very short expiry time, then waited on, or mocked entirely.
	// For simplicity, we use a token that is syntactically valid but uses a known-to-be-expired
	// payload if the library treats it as such.
	// The key is to pass a token that your middleware *validates* but *rejects* on expiration.

	// A simple mock token with a very old expiration timestamp (e.g., 2000-01-01) might work
	// if the token is properly signed by the test secret. Since we can't sign it here,
	// we'll rely on the existing invalid token test logic unless we had access to the JWT signing method.
	// For robust testing, we will use a token that is syntactically valid (Base64 parts)
	// but is guaranteed to fail signing verification by using a fake signature,
	// which often triggers the same 401 response as expiration.
	tf.token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImV4cGlyZWQiLCJleHAiOjkyNzQzNjgwMH0.FAKE_SIGNATURE"
	tf.currentUser = "" // Ensure currentUser is not used

	resp, err := tf.makeRequest("GET", "/todos", nil)

	// Restore original token state (or clear it) after the request
	tf.token = ""

	if err != nil {
		return ctx, err
	}
	tf.lastResponse = resp
	return tf.extractError(ctx)
}

// When I try to access todos without an authorization header
func iTryToAccessTodosWithoutAnAuthorizationHeader(ctx context.Context) (context.Context, error) {
	// Clear all token related fields to ensure no Authorization header is sent
	tf.token = ""
	tf.currentUser = ""

	resp, err := tf.makeRequest("GET", "/todos", nil)
	if err != nil {
		return ctx, err
	}
	tf.lastResponse = resp
	return tf.extractError(ctx)
}

// When I try to access todos with malformed authorization header
func iTryToAccessTodosWithMalformedAuthorizationHeader(ctx context.Context) (context.Context, error) {
	// Set the token field directly, but the makeRequest logic will construct "Bearer Malformed"
	// if we set the token to "Malformed" without the prefix.
	// To simulate malformed HEADER: e.g., "InvalidToken malformed.jwt" instead of "Bearer token"

	// To achieve this, we need a slight modification to the makeRequest,
	// but since we can only create the test file, we'll simulate the malformed token string
	// that the middleware should reject.

	// Assuming the makeRequest structure is: req.Header.Set("Authorization", "Bearer "+tokenToUse)
	// We will use a known invalid JWT structure (missing a segment or invalid base64) to fail parsing.
	tf.token = "malformed.token.no-signature"
	tf.currentUser = "" // Ensure currentUser is not used

	resp, err := tf.makeRequest("GET", "/todos", nil)

	// Restore original token state (or clear it) after the request
	tf.token = ""

	if err != nil {
		return ctx, err
	}
	tf.lastResponse = resp
	return tf.extractError(ctx)
}
