package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"todoapp/internal/app"
)

func theSecretKeyIsSetUp(ctx context.Context, secretKey string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}
	tc.Config.JWTSecret = secretKey
	return ctx, nil
}

func aUserNamedWithPasswordIsRegistered(ctx context.Context, username, password string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	creds := app.Credentials{Username: username, Password: password}
	resp, err := tc.MakeRequest("POST", "/register", creds)
	if err != nil {
		return ctx, fmt.Errorf("failed to register user: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		return ctx, fmt.Errorf("user registration failed with status %d", resp.StatusCode)
	}

	return ctx, nil
}

func userLogsInWithPasswordSuccessfully(ctx context.Context, username, password string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	creds := app.Credentials{Username: username, Password: password}
	// Use MakeRequest without setting CurrentUser first
	resp, err := tc.MakeRequest("POST", "/login", creds)
	if err != nil {
		return ctx, fmt.Errorf("failed to login user: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return ctx, fmt.Errorf("user login failed with status %d", resp.StatusCode)
	}

	var loginResponse map[string]string
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, fmt.Errorf("failed to read login response body: %w", err)
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &loginResponse); err != nil {
		return ctx, fmt.Errorf("failed to decode login response: %w (body: %s)", err, string(bodyBytes))
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

func theResponseStatusShouldBe(ctx context.Context, expectedStatus int) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp := tc.GetLastResponse()
	if resp == nil {
		return ctx, fmt.Errorf("no response available")
	}

	if resp.StatusCode != expectedStatus {
		// Read body for more context on error
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return ctx, fmt.Errorf("expected status %d, got %d (and failed to read body)", expectedStatus, resp.StatusCode)
		}
		err = resp.Body.Close()
		if err != nil {
			return nil, err
		}
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return ctx, fmt.Errorf("expected status %d, got %d. Body: %s", expectedStatus, resp.StatusCode, string(bodyBytes))
	}

	return ctx, nil
}

func userShouldReceiveASuccessMessage(ctx context.Context) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp := tc.GetLastResponse()
	if resp == nil {
		return ctx, fmt.Errorf("no response available")
	}

	var response map[string]string
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, fmt.Errorf("failed to read response body: %w", err)
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return ctx, fmt.Errorf("failed to decode response: %w (body: %s)", err, string(bodyBytes))
	}

	if _, exists := response["message"]; !exists {
		return ctx, fmt.Errorf("expected success message in response, got: %s", string(bodyBytes))
	}

	return ctx, nil
}

func userShouldReceiveAValidJwtToken(ctx context.Context, username string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp := tc.GetLastResponse()
	if resp == nil {
		return ctx, fmt.Errorf("no response available")
	}

	var loginResponse map[string]string
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, fmt.Errorf("failed to read response body: %w", err)
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &loginResponse); err != nil {
		return ctx, fmt.Errorf("failed to decode login response: %w (body: %s)", err, string(bodyBytes))
	}

	token, exists := loginResponse["token"]
	if !exists {
		return ctx, fmt.Errorf("no token in login response: %s", string(bodyBytes))
	}

	if username != "" {
		tc.StoreUserToken(username, token)
	}
	return ctx, nil
}

func userShouldReceiveAnErrorMessage(ctx context.Context, username, expectedError string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp := tc.GetLastResponse()
	if resp == nil {
		return ctx, fmt.Errorf("no response available")
	}

	var errorResponse map[string]string
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, fmt.Errorf("failed to read response body: %w", err)
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &errorResponse); err != nil {
		return ctx, fmt.Errorf("failed to decode error response: %w (body: %s)", err, string(bodyBytes))
	}

	actualError, exists := errorResponse["error"]
	if !exists {
		return ctx, fmt.Errorf("expected error message in response, got: %s", string(bodyBytes))
	}

	if actualError != expectedError {
		return ctx, fmt.Errorf("expected error '%s', got '%s'", expectedError, actualError)
	}

	return ctx, nil
}

func iShouldReceiveASuccessMessage(ctx context.Context) (context.Context, error) {
	return userShouldReceiveASuccessMessage(ctx)
}

func iShouldReceiveAValidJwtToken(ctx context.Context) (context.Context, error) {
	return userShouldReceiveAValidJwtToken(ctx, "")
}

func iShouldReceiveAnErrorMessage(ctx context.Context, expectedError string) (context.Context, error) {
	return userShouldReceiveAnErrorMessage(ctx, "", expectedError)
}
