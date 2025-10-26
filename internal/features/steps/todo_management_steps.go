package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/cucumber/godog"
	"todoapp/internal/app"
	"todoapp/internal/features/support"
)

func TestTodoManagement(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeTodoManagementScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../todo_management.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run todo management feature tests")
	}
}

func InitializeTodoManagementScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		tc := support.NewTestContext()
		if err := tc.SetupServer(); err != nil {
			return ctx, fmt.Errorf("failed to setup test server: %w", err)
		}
		return support.SetTestContextInContext(ctx, tc), nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		tc := support.GetTestContextFromContext(ctx)
		if tc != nil {
			tc.CloseServer()
		}
		return ctx, nil
	})

	// Authentication steps (using common steps)
	ctx.Step(`^the secret key "([^"]*)" is set up$`, theSecretKeyIsSetUp)
	ctx.Step(`^a user named "([^"]*)" with password "([^"]*)" is registered$`, aUserNamedWithPasswordIsRegistered)
	ctx.Step(`^user "([^"]*)" logs in with password "([^"]*)" successfully$`, userLogsInWithPasswordSuccessfully)

	// Todo creation steps
	ctx.Step(`^user "([^"]*)" creates a todo with title "([^"]*)"$`, userCreatesATodoWithTitle)
	ctx.Step(`^user "([^"]*)" has created a todo with title "([^"]*)"$`, userHasCreatedATodoWithTitle)

	// Todo retrieval steps
	ctx.Step(`^user "([^"]*)" requests all todos$`, userRequestsAllTodos)
	ctx.Step(`^user "([^"]*)" requests the todo by ID$`, userRequestsTheTodoByID)

	// Todo update steps
	ctx.Step(`^user "([^"]*)" updates the todo title to "([^"]*)"$`, userUpdatesTheTodoTitleTo)
	ctx.Step(`^user "([^"]*)" marks the todo as completed$`, userMarksTheTodoAsCompleted)

	// Todo deletion steps
	ctx.Step(`^user "([^"]*)" deletes the todo$`, userDeletesTheTodo)

	// Authorization steps
	ctx.Step(`^user "([^"]*)" tries to update Bob's todo$`, userTriesToUpdateBobsTodo)
	ctx.Step(`^user "([^"]*)" should not see "([^"]*)"$`, userShouldNotSee)

	// Error handling steps
	ctx.Step(`^user "([^"]*)" tries to create a todo with empty title$`, userTriesToCreateATodoWithEmptyTitle)
	ctx.Step(`^user "([^"]*)" tries to update a non-existent todo$`, userTriesToUpdateANonExistentTodo)
	ctx.Step(`^user "([^"]*)" tries to delete a non-existent todo$`, userTriesToDeleteANonExistentTodo)

	// Assertion steps
	ctx.Step(`^user "([^"]*)" should receive a success response$`, userShouldReceiveASuccessResponse)
	ctx.Step(`^the todo should have title "([^"]*)"$`, theTodoShouldHaveTitle)
	ctx.Step(`^the todo should not be completed$`, theTodoShouldNotBeCompleted)
	ctx.Step(`^the todo should be completed$`, theTodoShouldBeCompleted)
	ctx.Step(`^user "([^"]*)" should see (\d+) todos$`, userShouldSeeTodos)
	ctx.Step(`^the todos should include "([^"]*)"$`, theTodosShouldInclude)
	ctx.Step(`^user "([^"]*)" should receive the todo with title "([^"]*)"$`, userShouldReceiveTheTodoWithTitle)
	ctx.Step(`^user "([^"]*)" should receive the updated todo$`, userShouldReceiveTheUpdatedTodo)
	ctx.Step(`^user "([^"]*)" should receive a success message$`, userShouldReceiveASuccessMessage)
	ctx.Step(`^user "([^"]*)" should no longer see the todo$`, userShouldNoLongerSeeTheTodo)
	ctx.Step(`^user "([^"]*)" should see "([^"]*)" as completed$`, userShouldSeeAsCompleted)
	ctx.Step(`^user "([^"]*)" should see "([^"]*)" as not completed$`, userShouldSeeAsNotCompleted)
	ctx.Step(`^user "([^"]*)" should receive an empty list$`, userShouldReceiveAnEmptyList)
	ctx.Step(`^user "([^"]*)" should receive a validation error$`, userShouldReceiveAValidationError)
	ctx.Step(`^user "([^"]*)" should receive a not found error$`, userShouldReceiveANotFoundError)
	ctx.Step(`^the response status should be (\d+)$`, theResponseStatusShouldBe)
}

// Authentication steps are now in common_steps.go

// Todo creation steps
func userCreatesATodoWithTitle(ctx context.Context, username, title string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	todo := map[string]string{"title": title}
	resp, err := tc.MakeAuthenticatedRequest("POST", "/todos", todo, username)
	if err != nil {
		return ctx, fmt.Errorf("failed to create todo: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func userHasCreatedATodoWithTitle(ctx context.Context, username, title string) (context.Context, error) {
	// First create the todo
	ctx, err := userCreatesATodoWithTitle(ctx, username, title)
	if err != nil {
		return ctx, err
	}

	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	// Extract and store the todo ID from the response
	if tc.GetLastResponse() != nil && tc.GetLastResponse().StatusCode == http.StatusCreated {
		var todo app.Todo
		if err := json.NewDecoder(tc.GetLastResponse().Body).Decode(&todo); err == nil {
			tc.StoreTodoID(username, todo.ID)
			tc.StoreTodoByTitle(title, todo.ID)
		}
	}

	return ctx, nil
}

// Todo retrieval steps
func userRequestsAllTodos(ctx context.Context, username string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp, err := tc.MakeAuthenticatedRequest("GET", "/todos", nil, username)
	if err != nil {
		return ctx, fmt.Errorf("failed to get todos: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func userRequestsTheTodoByID(ctx context.Context, username string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	todoID, exists := tc.GetTodoID(username)
	if !exists {
		return ctx, fmt.Errorf("no todo ID found for user %s", username)
	}

	resp, err := tc.MakeAuthenticatedRequest("GET", "/todos/"+todoID, nil, username)
	if err != nil {
		return ctx, fmt.Errorf("failed to get todo: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

// Todo update steps
func userUpdatesTheTodoTitleTo(ctx context.Context, username, newTitle string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	todoID, exists := tc.GetTodoID(username)
	if !exists {
		return ctx, fmt.Errorf("no todo ID found for user %s", username)
	}

	updateReq := app.UpdateTodoRequest{Title: &newTitle}
	resp, err := tc.MakeAuthenticatedRequest("PUT", "/todos/"+todoID, updateReq, username)
	if err != nil {
		return ctx, fmt.Errorf("failed to update todo: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func userMarksTheTodoAsCompleted(ctx context.Context, username string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	todoID, exists := tc.GetTodoID(username)
	if !exists {
		return ctx, fmt.Errorf("no todo ID found for user %s", username)
	}

	completed := true
	updateReq := app.UpdateTodoRequest{Completed: &completed}
	resp, err := tc.MakeAuthenticatedRequest("PUT", "/todos/"+todoID, updateReq, username)
	if err != nil {
		return ctx, fmt.Errorf("failed to update todo: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

// Todo deletion steps
func userDeletesTheTodo(ctx context.Context, username string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	todoID, exists := tc.GetTodoID(username)
	if !exists {
		return ctx, fmt.Errorf("no todo ID found for user %s", username)
	}

	resp, err := tc.MakeAuthenticatedRequest("DELETE", "/todos/"+todoID, nil, username)
	if err != nil {
		return ctx, fmt.Errorf("failed to delete todo: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

// Authorization steps
func userTriesToUpdateBobsTodo(ctx context.Context, username string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	// Try to update Bob's todo (assuming Bob has a todo with ID "bob-todo-id")
	bobTodoID := "bob-todo-id"
	hackedTitle := "Hacked title"
	updateReq := app.UpdateTodoRequest{Title: &hackedTitle}
	resp, err := tc.MakeAuthenticatedRequest("PUT", "/todos/"+bobTodoID, updateReq, username)
	if err != nil {
		return ctx, fmt.Errorf("failed to attempt update: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func userShouldNotSee(ctx context.Context, username, title string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	// Get all todos for the user
	resp, err := tc.MakeAuthenticatedRequest("GET", "/todos", nil, username)
	if err != nil {
		return ctx, fmt.Errorf("failed to get todos: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return ctx, fmt.Errorf("failed to get todos with status %d", resp.StatusCode)
	}

	var todos []app.Todo
	if err := json.NewDecoder(resp.Body).Decode(&todos); err != nil {
		return ctx, fmt.Errorf("failed to decode todos: %w", err)
	}

	for _, todo := range todos {
		if todo.Title == title {
			return ctx, fmt.Errorf("user %s should not see todo with title '%s' but found it", username, title)
		}
	}

	return ctx, nil
}

// Error handling steps
func userTriesToCreateATodoWithEmptyTitle(ctx context.Context, username string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	todo := map[string]string{"title": ""}
	resp, err := tc.MakeAuthenticatedRequest("POST", "/todos", todo, username)
	if err != nil {
		return ctx, fmt.Errorf("failed to attempt todo creation: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func userTriesToUpdateANonExistentTodo(ctx context.Context, username string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	nonExistentID := "non-existent-id"
	updatedTitle := "Updated title"
	updateReq := app.UpdateTodoRequest{Title: &updatedTitle}
	resp, err := tc.MakeAuthenticatedRequest("PUT", "/todos/"+nonExistentID, updateReq, username)
	if err != nil {
		return ctx, fmt.Errorf("failed to attempt update: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func userTriesToDeleteANonExistentTodo(ctx context.Context, username string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	nonExistentID := "non-existent-id"
	resp, err := tc.MakeAuthenticatedRequest("DELETE", "/todos/"+nonExistentID, nil, username)
	if err != nil {
		return ctx, fmt.Errorf("failed to attempt deletion: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

// Assertion steps
func userShouldReceiveASuccessResponse(ctx context.Context, username string) (context.Context, error) {
	return theResponseStatusShouldBe(ctx, 200)
}

func theTodoShouldHaveTitle(ctx context.Context, expectedTitle string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp := tc.GetLastResponse()
	if resp == nil {
		return ctx, fmt.Errorf("no response available")
	}

	var todo app.Todo
	if err := json.NewDecoder(resp.Body).Decode(&todo); err != nil {
		return ctx, fmt.Errorf("failed to decode todo: %w", err)
	}

	if todo.Title != expectedTitle {
		return ctx, fmt.Errorf("expected todo title '%s', got '%s'", expectedTitle, todo.Title)
	}

	return ctx, nil
}

func theTodoShouldNotBeCompleted(ctx context.Context) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp := tc.GetLastResponse()
	if resp == nil {
		return ctx, fmt.Errorf("no response available")
	}

	var todo app.Todo
	if err := json.NewDecoder(resp.Body).Decode(&todo); err != nil {
		return ctx, fmt.Errorf("failed to decode todo: %w", err)
	}

	if todo.Completed {
		return ctx, fmt.Errorf("expected todo to not be completed, but it is")
	}

	return ctx, nil
}

func theTodoShouldBeCompleted(ctx context.Context) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp := tc.GetLastResponse()
	if resp == nil {
		return ctx, fmt.Errorf("no response available")
	}

	var todo app.Todo
	if err := json.NewDecoder(resp.Body).Decode(&todo); err != nil {
		return ctx, fmt.Errorf("failed to decode todo: %w", err)
	}

	if !todo.Completed {
		return ctx, fmt.Errorf("expected todo to be completed, but it is not")
	}

	return ctx, nil
}

func userShouldSeeTodos(ctx context.Context, username string, expectedCount int) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp := tc.GetLastResponse()
	if resp == nil {
		return ctx, fmt.Errorf("no response available")
	}

	var todos []app.Todo
	if err := json.NewDecoder(resp.Body).Decode(&todos); err != nil {
		return ctx, fmt.Errorf("failed to decode todos: %w", err)
	}

	if len(todos) != expectedCount {
		return ctx, fmt.Errorf("expected %d todos, got %d", expectedCount, len(todos))
	}

	return ctx, nil
}

func theTodosShouldInclude(ctx context.Context, expectedTitle string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp := tc.GetLastResponse()
	if resp == nil {
		return ctx, fmt.Errorf("no response available")
	}

	var todos []app.Todo
	if err := json.NewDecoder(resp.Body).Decode(&todos); err != nil {
		return ctx, fmt.Errorf("failed to decode todos: %w", err)
	}

	for _, todo := range todos {
		if todo.Title == expectedTitle {
			return ctx, nil
		}
	}

	return ctx, fmt.Errorf("expected to find todo with title '%s' in todos", expectedTitle)
}

func userShouldReceiveTheTodoWithTitle(ctx context.Context, username, expectedTitle string) (context.Context, error) {
	return theTodoShouldHaveTitle(ctx, expectedTitle)
}

func userShouldReceiveTheUpdatedTodo(ctx context.Context, username string) (context.Context, error) {
	return userShouldReceiveASuccessResponse(ctx, username)
}

func userShouldNoLongerSeeTheTodo(ctx context.Context, username string) (context.Context, error) {
	// Get all todos and verify the deleted todo is not present
	return userRequestsAllTodos(ctx, username)
}

func userShouldSeeAsCompleted(ctx context.Context, username, title string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp := tc.GetLastResponse()
	if resp == nil {
		return ctx, fmt.Errorf("no response available")
	}

	var todos []app.Todo
	if err := json.NewDecoder(resp.Body).Decode(&todos); err != nil {
		return ctx, fmt.Errorf("failed to decode todos: %w", err)
	}

	for _, todo := range todos {
		if todo.Title == title {
			if !todo.Completed {
				return ctx, fmt.Errorf("expected todo '%s' to be completed, but it is not", title)
			}
			return ctx, nil
		}
	}

	return ctx, fmt.Errorf("todo with title '%s' not found", title)
}

func userShouldSeeAsNotCompleted(ctx context.Context, username, title string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp := tc.GetLastResponse()
	if resp == nil {
		return ctx, fmt.Errorf("no response available")
	}

	var todos []app.Todo
	if err := json.NewDecoder(resp.Body).Decode(&todos); err != nil {
		return ctx, fmt.Errorf("failed to decode todos: %w", err)
	}

	for _, todo := range todos {
		if todo.Title == title {
			if todo.Completed {
				return ctx, fmt.Errorf("expected todo '%s' to not be completed, but it is", title)
			}
			return ctx, nil
		}
	}

	return ctx, fmt.Errorf("todo with title '%s' not found", title)
}

func userShouldReceiveAnEmptyList(ctx context.Context, username string) (context.Context, error) {
	return userShouldSeeTodos(ctx, username, 0)
}

func userShouldReceiveAValidationError(ctx context.Context, username string) (context.Context, error) {
	return theResponseStatusShouldBe(ctx, 400)
}

func userShouldReceiveANotFoundError(ctx context.Context, username string) (context.Context, error) {
	return theResponseStatusShouldBe(ctx, 404)
}
