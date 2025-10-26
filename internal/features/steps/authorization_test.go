package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"net/http"
	"time"
	"todoapp/internal/app"
)

func userGetsAllTodos(ctx context.Context, username string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}
	tc.SetCurrentUser(username)

	resp, err := tc.MakeRequest("GET", "/todos", nil)
	if err != nil {
		return ctx, err
	}
	tc.SetLastResponse(resp)
	return ctx, nil
}

func todoTitleExistsInLastResponse(tc *TestContext, title string) (bool, error) {
	resp := tc.GetLastResponse()
	if resp == nil {
		return false, fmt.Errorf("no response to check")
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %w", err)
	}
	err = resp.Body.Close()
	if err != nil {
		return false, err
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var todos []app.Todo
	if err := json.Unmarshal(bodyBytes, &todos); err != nil {
		if resp.StatusCode != http.StatusOK {
			return false, nil
		}
		return false, fmt.Errorf("failed to decode response body as []app.Todo: %w (body: %s)", err, string(bodyBytes))
	}

	for _, todo := range todos {
		if todo.Title == title {
			return true, nil
		}
	}

	return false, nil
}

func userShouldOnlySee(ctx context.Context, username, title string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	found, err := todoTitleExistsInLastResponse(tc, title)
	if err != nil {
		return ctx, err
	}
	if !found {
		return ctx, fmt.Errorf("user %s expected to see todo '%s' but did not", username, title)
	}
	return ctx, nil
}

func userShouldNotSee(ctx context.Context, username, title string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	found, err := todoTitleExistsInLastResponse(tc, title)

	if err != nil {
		return ctx, err
	}
	if found {
		return ctx, fmt.Errorf("user %s expected NOT to see todo '%s' but DID", username, title)
	}
	return ctx, nil
}

func userTriesToGetAnotherUsersTodoByID(ctx context.Context, requestingUser, ownerUser string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}
	tc.SetCurrentUser(requestingUser)

	var ownerTodoTitle string
	if ownerUser == "bob" {
		ownerTodoTitle = "Bob's Private Task"
	} else {
		ownerTodoTitle = "Alice's Private Task"
	}

	todoID, ok := tc.GetTodoIDByTitle(ownerTodoTitle)
	if !ok {
		return ctx, fmt.Errorf("could not find todo ID for title '%s' (owner: %s)", ownerTodoTitle, ownerUser)
	}

	resp, err := tc.MakeRequest("GET", "/todos/"+todoID, nil)
	if err != nil {
		return ctx, err
	}
	tc.SetLastResponse(resp)
	return ctx, nil
}

func userTriesToUpdateAnotherUsersTodo(ctx context.Context, requestingUser, ownerUser string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}
	tc.SetCurrentUser(requestingUser)

	todoID, ok := tc.GetTodoIDByTitle("Alice's Task")
	if !ok {
		return ctx, fmt.Errorf("could not find todo ID for title 'Alice's Task' (owner: %s)", ownerUser)
	}

	newTitle := "Attempted update"
	reqBody := app.UpdateTodoRequest{Title: &newTitle}

	resp, err := tc.MakeRequest("PUT", "/todos/"+todoID, reqBody)
	if err != nil {
		return ctx, err
	}
	tc.SetLastResponse(resp)
	return ctx, nil
}

func userTriesToDeleteAnotherUsersTodo(ctx context.Context, requestingUser, ownerUser string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}
	tc.SetCurrentUser(requestingUser)

	todoID, ok := tc.GetTodoIDByTitle("Alice's Task")
	if !ok {
		return ctx, fmt.Errorf("could not find todo ID for title 'Alice's Task' (owner: %s)", ownerUser)
	}

	resp, err := tc.MakeRequest("DELETE", "/todos/"+todoID, nil)
	if err != nil {
		return ctx, err
	}
	tc.SetLastResponse(resp)
	return ctx, nil
}

func makeRequestWithToken(ctx context.Context, tokenValue string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	tc.SetCurrentUser("")

	req, err := http.NewRequestWithContext(context.Background(), "GET", tc.BaseURL+"/todos", nil)
	if err != nil {
		return ctx, err
	}

	switch tokenValue {
	case "MISSING":
	case "MALFORMED":
		req.Header.Set("Authorization", "Bearer.malformed.token")
	default:
		req.Header.Set("Authorization", "Bearer "+tokenValue)
	}

	resp, err := tc.Client.Do(req)
	if err != nil {
		return ctx, err
	}
	tc.SetLastResponse(resp)
	return ctx, nil
}

func iTryToAccessTodosWithAnInvalidToken(ctx context.Context) (context.Context, error) {
	invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFsaWNlIn0.INVALID_SIGNATURE_HERE"
	return makeRequestWithToken(ctx, invalidToken)
}

func iTryToAccessTodosWithAnExpiredToken(ctx context.Context) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	claims := jwt.MapClaims{
		"username": "test",
		"exp":      time.Now().Add(-1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tc.Config.JWTSecret))
	if err != nil {
		return ctx, fmt.Errorf("failed to create expired token: %w", err)
	}

	return makeRequestWithToken(ctx, tokenString)
}

func iTryToAccessTodosWithoutAnAuthorizationHeader(ctx context.Context) (context.Context, error) {
	return makeRequestWithToken(ctx, "MISSING")
}

func iTryToAccessTodosWithMalformedAuthorizationHeader(ctx context.Context) (context.Context, error) {
	return makeRequestWithToken(ctx, "MALFORMED")
}
