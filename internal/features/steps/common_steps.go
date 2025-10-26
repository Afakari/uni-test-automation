package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"todoapp/internal/app"
	"todoapp/internal/features/support"
)

func theSecretKeyIsSetUp(ctx context.Context, secretKey string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}
	tc.Config.JWTSecret = secretKey
	return ctx, nil
}

func aUserNamedWithPasswordIsRegistered(ctx context.Context, username, password string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	creds := app.Credentials{Username: username, Password: password}
	resp, err := tc.MakeRequest("POST", "/register", creds)
	if err != nil {
		return ctx, fmt.Errorf("failed to register user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		return ctx, fmt.Errorf("user registration failed with status %d", resp.StatusCode)
	}

	return ctx, nil
}

func userLogsInWithPasswordSuccessfully(ctx context.Context, username, password string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	creds := app.Credentials{Username: username, Password: password}
	resp, err := tc.MakeRequest("POST", "/login", creds)
	if err != nil {
		return ctx, fmt.Errorf("failed to login user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ctx, fmt.Errorf("user login failed with status %d", resp.StatusCode)
	}

	var loginResponse map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return ctx, fmt.Errorf("failed to decode login response: %w", err)
	}

	token, exists := loginResponse["token"]
	if !exists {
		return ctx, fmt.Errorf("no token in login response")
	}

	tc.StoreUserToken(username, token)
	tc.SetCurrentUser(username)
	tc.SetLastResponse(resp)

	return ctx, nil
}

// Common assertion steps
func theResponseStatusShouldBe(ctx context.Context, expectedStatus int) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp := tc.GetLastResponse()
	if resp == nil {
		return ctx, fmt.Errorf("no response available")
	}

	if resp.StatusCode != expectedStatus {
		return ctx, fmt.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	return ctx, nil
}

func userShouldReceiveASuccessMessage(ctx context.Context, username string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp := tc.GetLastResponse()
	if resp == nil {
		return ctx, fmt.Errorf("no response available")
	}

	var response map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return ctx, fmt.Errorf("failed to decode response: %w", err)
	}

	if _, exists := response["message"]; !exists {
		return ctx, fmt.Errorf("expected success message in response")
	}

	return ctx, nil
}

func userShouldReceiveAValidJwtToken(ctx context.Context, username string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp := tc.GetLastResponse()
	if resp == nil {
		return ctx, fmt.Errorf("no response available")
	}

	var loginResponse map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return ctx, fmt.Errorf("failed to decode login response: %w", err)
	}

	token, exists := loginResponse["token"]
	if !exists {
		return ctx, fmt.Errorf("no token in login response")
	}

	tc.StoreUserToken(username, token)
	return ctx, nil
}

func userShouldReceiveAnErrorMessage(ctx context.Context, username, expectedError string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp := tc.GetLastResponse()
	if resp == nil {
		return ctx, fmt.Errorf("no response available")
	}

	var errorResponse map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
		return ctx, fmt.Errorf("failed to decode error response: %w", err)
	}

	actualError, exists := errorResponse["error"]
	if !exists {
		return ctx, fmt.Errorf("expected error message in response")
	}

	if actualError != expectedError {
		return ctx, fmt.Errorf("expected error '%s', got '%s'", expectedError, actualError)
	}

	return ctx, nil
}
