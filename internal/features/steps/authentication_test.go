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

func iRegisterWithUsernameAndPassword(ctx context.Context, username, password string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	creds := app.Credentials{Username: username, Password: password}
	resp, err := tc.MakeRequest("POST", "/register", creds)
	if err != nil {
		return ctx, fmt.Errorf("failed to register user: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func iLoginWithUsernameAndPassword(ctx context.Context, username, password string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	creds := app.Credentials{Username: username, Password: password}
	resp, err := tc.MakeRequest("POST", "/login", creds)
	if err != nil {
		return ctx, fmt.Errorf("failed to login user: %w", err)
	}

	tc.SetLastResponse(resp)

	if resp.StatusCode == http.StatusOK {
		var loginResponse map[string]string
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return ctx, err
		}
		err = resp.Body.Close()
		if err != nil {
			return nil, err
		}
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if err := json.Unmarshal(bodyBytes, &loginResponse); err == nil {
			if token, exists := loginResponse["token"]; exists {
				tc.StoreUserToken(username, token)
				tc.SetCurrentUser(username)
			}
		}
	}
	return ctx, nil
}

func iSendInvalidJSONToTheRegisterEndpoint(ctx context.Context) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp, err := tc.MakeRequest("POST", "/register", "not a json string")
	if err != nil {
		return ctx, fmt.Errorf("failed to send invalid JSON: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func iSendInvalidJSONToTheLoginEndpoint(ctx context.Context) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp, err := tc.MakeRequest("POST", "/login", "not a json string")
	if err != nil {
		return ctx, fmt.Errorf("failed to send invalid JSON: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}
