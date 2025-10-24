package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/cucumber/godog"
	"todoapp/internal/app"
)

// --- Feature Runner (TestUserAuthentication) ---

func TestUserAuthentication(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeUserAuthScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../authentication.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeUserAuthScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		tf = &todoFeature{}
		ctx, _ = tf.setupServer(ctx)
		return ctx, nil
	})
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		ctx, err = tf.closeSever(ctx)
		return ctx, err
	})

	ctx.Step(`^the secret key "([^"]*)" is set up$`, isJwtSecretSet)
	ctx.Step(`^a user named "([^"]*)" with password "([^"]*)" is already registered$`, aUserNamedAliceWithPasswordPassIsRegistered)

	ctx.Step(`^I register with username "([^"]*)" and password "([^"]*)"$`, iRegisterWithUsernameAndPassword)
	ctx.Step(`^I login with username "([^"]*)" and password "([^"]*)"$`, iLoginWithUsernameAndPassword)
	ctx.Step(`^I send invalid JSON to the register endpoint$`, iSendInvalidJSONToTheRegisterEndpoint)
	ctx.Step(`^I send invalid JSON to the login endpoint$`, iSendInvalidJSONToTheLoginEndpoint)
	ctx.Step(`^a user named "([^"]*)" with password "([^"]*)" is registered$`, aUserNamedWithPasswordIsRegistered)

	ctx.Step(`^I should receive a success message$`, iShouldReceiveASuccessMessage)
	ctx.Step(`^I should receive a valid JWT token$`, iShouldReceiveAValidJwtToken)
	ctx.Step(`^I should receive an error message "([^"]*)"$`, iShouldReceiveAnErrorMessage)
	ctx.Step(`^the response status should be (\d+)$`, theResponseStatusShouldBe)
}

func iRegisterWithUsernameAndPassword(ctx context.Context, username, password string) (context.Context, error) {
	body := app.Credentials{Username: username, Password: password}

	resp, err := tf.makeRequest("POST", "/register", body)
	if err != nil {
		return ctx, err
	}
	tf.lastResponse = resp
	return tf.extractError(ctx)
}

func iLoginWithUsernameAndPassword(ctx context.Context, username, password string) (context.Context, error) {
	body := app.Credentials{Username: username, Password: password}
	resp, err := tf.makeRequest("POST", "/login", body)
	if err != nil {
		return ctx, err
	}
	tf.lastResponse = resp
	return tf.extractError(ctx)
}

func iSendInvalidJSONToTheRegisterEndpoint(ctx context.Context) (context.Context, error) {
	resp, err := tf.makeRequest("POST", "/register", "not a json string")
	if err != nil {
		return ctx, err
	}
	tf.lastResponse = resp
	return tf.extractError(ctx)
}

func iSendInvalidJSONToTheLoginEndpoint(ctx context.Context) (context.Context, error) {
	resp, err := tf.makeRequest("POST", "/login", "not a json string")
	if err != nil {
		return ctx, err
	}
	tf.lastResponse = resp
	return tf.extractError(ctx)
}

func iShouldReceiveASuccessMessage(ctx context.Context) (context.Context, error) {
	// Ensure we have a response to check
	if tf.lastResponse == nil {
		return ctx, fmt.Errorf("no request was made, tf.lastResponse is nil")
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil { /* Handle error if needed */
		}
	}(tf.lastResponse.Body)

	var resBody map[string]string
	if err := json.NewDecoder(tf.lastResponse.Body).Decode(&resBody); err != nil {
		return ctx, fmt.Errorf("failed to decode response body: %w", err)
	}

	if _, ok := resBody["message"]; !ok {
		return ctx, fmt.Errorf("expected a success message, but didn't find 'message' in response: %v", resBody)
	}
	return ctx, nil
}

func iShouldReceiveAValidJwtToken(ctx context.Context) (context.Context, error) {
	if tf.lastResponse == nil {
		return ctx, fmt.Errorf("no request was made, tf.lastResponse is nil")
	}

	// This defers closing the body, even if an error occurs below.
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil { /* Handle error if needed */
		}
	}(tf.lastResponse.Body)

	var resBody map[string]string
	if err := json.NewDecoder(tf.lastResponse.Body).Decode(&resBody); err != nil {
		return ctx, fmt.Errorf("failed to decode response body: %w", err)
	}

	tf.token = resBody["token"] // Store the token for subsequent protected steps
	if tf.token == "" {
		return ctx, fmt.Errorf("login failed, no token returned in response: %v", resBody)
	}
	return ctx, nil
}

func iShouldReceiveAnErrorMessage(ctx context.Context, expected string) (context.Context, error) {
	if tf.lastErrorMsg == "" {
		return ctx, fmt.Errorf("no error message was extracted from the response body")
	}

	if tf.lastErrorMsg != expected {
		return ctx, fmt.Errorf("expected error message '%s' but got '%s'", expected, tf.lastErrorMsg)
	}
	return ctx, nil
}

func theResponseStatusShouldBe(ctx context.Context, expectedStatus int) (context.Context, error) {
	if tf.lastResponse == nil {
		return ctx, fmt.Errorf("no request was made, tf.lastResponse is nil")
	}

	if tf.lastResponse.StatusCode != expectedStatus {
		return ctx, fmt.Errorf("expected status code %d, but got %d", expectedStatus, tf.lastResponse.StatusCode)
	}
	return ctx, nil
}

// Helper to extract the error message from the response body for error scenarios
// Made this a method on todoFeature to access tf.lastResponse and tf.lastErrorMsg
func (tf *todoFeature) extractError(ctx context.Context) (context.Context, error) {
	resp := tf.lastResponse

	// Reset error message state for the next check
	tf.lastErrorMsg = ""

	// Only attempt to read/decode the body if it's an error status
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {

		// IMPORTANT: We need a way to read the body and then immediately close it,
		// but not close it before the assertion step reads it.
		// Since extractError is called immediately after the request, we will
		// read the body into a byte buffer and replace the response body with
		// a new reader so the assertion step can read it again.
		// This is a common pattern for inspecting response bodies in middleware/testing.

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return ctx, fmt.Errorf("failed to read response body for error extraction: %w", err)
		}

		// Reset the response body so subsequent assertion steps can read it
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var resBody map[string]string
		// Attempt to decode a gin.H{"error": "..."} response
		if err := json.Unmarshal(bodyBytes, &resBody); err == nil {
			if errMsg, ok := resBody["error"]; ok {
				tf.lastErrorMsg = errMsg
			}
		}
	}

	return ctx, nil
}

func aUserNamedWithPasswordIsRegistered(username, password string) error {
	body := app.Credentials{Username: username, Password: password}

	resp, err := tf.makeRequest("POST", "/register", body)
	if err != nil {
		return fmt.Errorf("failed to make registration request: %w", err)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("user registration failed, expected status 201, got %v", resp.StatusCode)
	}

	return nil
}
