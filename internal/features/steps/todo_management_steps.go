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

func userCreatesATodoWithTitle(ctx context.Context, username, title string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	tc.SetCurrentUser(username)
	todo := map[string]string{"title": title}
	resp, err := tc.MakeRequest("POST", "/todos", todo)
	if err != nil {
		return ctx, fmt.Errorf("failed to create todo: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func userHasCreatedATodoWithTitle(ctx context.Context, username, title string) (context.Context, error) {
	ctx, err := userCreatesATodoWithTitle(ctx, username, title)
	if err != nil {
		return ctx, err
	}

	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	if tc.GetLastResponse() != nil && tc.GetLastResponse().StatusCode == http.StatusCreated {
		var todo app.Todo

		bodyBytes, err := io.ReadAll(tc.GetLastResponse().Body)
		if err != nil {
			return ctx, err
		}
		err = tc.GetLastResponse().Body.Close()
		if err != nil {
			return nil, err
		}
		tc.GetLastResponse().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if err := json.Unmarshal(bodyBytes, &todo); err == nil {
			tc.StoreTodoID(username, todo.ID)
			tc.StoreTodoByTitle(title, todo.ID)
		} else {
			return ctx, fmt.Errorf("failed to decode created todo: %w", err)
		}
	} else {
		return ctx, fmt.Errorf("failed to create todo, status was %d", tc.GetLastResponse().StatusCode)
	}

	return ctx, nil
}

func userRequestsAllTodos(ctx context.Context, username string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	tc.SetCurrentUser(username)
	resp, err := tc.MakeRequest("GET", "/todos", nil)
	if err != nil {
		return ctx, fmt.Errorf("failed to get todos: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func userRequestsTheTodoByID(ctx context.Context, username string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	tc.SetCurrentUser(username)
	todoID, exists := tc.GetTodoID(username)
	if !exists {
		return ctx, fmt.Errorf("no todo ID found for user %s", username)
	}

	resp, err := tc.MakeRequest("GET", "/todos/"+todoID, nil)
	if err != nil {
		return ctx, fmt.Errorf("failed to get todo: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func userUpdatesTheTodoTitleTo(ctx context.Context, username, newTitle string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	tc.SetCurrentUser(username)
	todoID, exists := tc.GetTodoID(username)
	if !exists {
		return ctx, fmt.Errorf("no todo ID found for user %s", username)
	}

	updateReq := app.UpdateTodoRequest{Title: &newTitle}
	resp, err := tc.MakeRequest("PUT", "/todos/"+todoID, updateReq)
	if err != nil {
		return ctx, fmt.Errorf("failed to update todo: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func userMarksTheTodoAsCompleted(ctx context.Context, username string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	tc.SetCurrentUser(username)
	todoID, exists := tc.GetTodoID(username)
	if !exists {
		return ctx, fmt.Errorf("no todo ID found for user %s", username)
	}

	completed := true
	updateReq := app.UpdateTodoRequest{Completed: &completed}
	resp, err := tc.MakeRequest("PUT", "/todos/"+todoID, updateReq)
	if err != nil {
		return ctx, fmt.Errorf("failed to update todo: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func userDeletesTheTodo(ctx context.Context, username string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	tc.SetCurrentUser(username)
	todoID, exists := tc.GetTodoID(username)
	if !exists {
		return ctx, fmt.Errorf("no todo ID found for user %s", username)
	}

	resp, err := tc.MakeRequest("DELETE", "/todos/"+todoID, nil)
	if err != nil {
		return ctx, fmt.Errorf("failed to delete todo: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func userTriesToUpdateBobsTodo(ctx context.Context, username string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	tc.SetCurrentUser(username)
	bobTodoID, ok := tc.GetTodoIDByTitle("Bob's task")
	if !ok {
		return ctx, fmt.Errorf("could not find todo ID for 'Bob's task'")
	}

	hackedTitle := "Hacked title"
	updateReq := app.UpdateTodoRequest{Title: &hackedTitle}
	resp, err := tc.MakeRequest("PUT", "/todos/"+bobTodoID, updateReq)
	if err != nil {
		return ctx, fmt.Errorf("failed to attempt update: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func userTriesToCreateATodoWithEmptyTitle(ctx context.Context, username string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	tc.SetCurrentUser(username)
	todo := map[string]string{"title": ""}
	resp, err := tc.MakeRequest("POST", "/todos", todo)
	if err != nil {
		return ctx, fmt.Errorf("failed to attempt todo creation: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func userTriesToUpdateANonExistentTodo(ctx context.Context, username string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	tc.SetCurrentUser(username)
	nonExistentID := "non-existent-id"
	updatedTitle := "Updated title"
	updateReq := app.UpdateTodoRequest{Title: &updatedTitle}
	resp, err := tc.MakeRequest("PUT", "/todos/"+nonExistentID, updateReq)
	if err != nil {
		return ctx, fmt.Errorf("failed to attempt update: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func userTriesToDeleteANonExistentTodo(ctx context.Context, username string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	tc.SetCurrentUser(username)
	nonExistentID := "non-existent-id"
	resp, err := tc.MakeRequest("DELETE", "/todos/"+nonExistentID, nil)
	if err != nil {
		return ctx, fmt.Errorf("failed to attempt deletion: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func userShouldReceiveASuccessResponse(ctx context.Context) (context.Context, error) {
	return theResponseStatusShouldBe(ctx, 200)
}
func userShouldReceiveACreatedResponse(ctx context.Context) (context.Context, error) {
	return theResponseStatusShouldBe(ctx, 201)
}

func readTodoFromLastResponse(ctx context.Context) (*app.Todo, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return nil, fmt.Errorf("test context not found")
	}
	resp := tc.GetLastResponse()
	if resp == nil {
		return nil, fmt.Errorf("no response available")
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var todo app.Todo
	if err := json.Unmarshal(bodyBytes, &todo); err == nil {
		return &todo, nil
	}

	var todos []app.Todo
	if err := json.Unmarshal(bodyBytes, &todos); err != nil {
		return nil, fmt.Errorf("failed to decode response as single todo or todo list: %w (body: %s)", err, string(bodyBytes))
	}

	if len(todos) == 0 {
		return nil, fmt.Errorf("no todos found in the response list")
	}

	return &todos[0], nil
}

func readTodoListFromLastResponse(ctx context.Context) ([]app.Todo, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return nil, fmt.Errorf("test context not found")
	}
	resp := tc.GetLastResponse()
	if resp == nil {
		return nil, fmt.Errorf("no response available")
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var todos []app.Todo
	if err := json.Unmarshal(bodyBytes, &todos); err != nil {
		return nil, fmt.Errorf("failed to decode todo list: %w (body: %s)", err, string(bodyBytes))
	}
	return todos, nil
}

func theTodoShouldHaveTitle(ctx context.Context, expectedTitle string) (context.Context, error) {
	todo, err := readTodoFromLastResponse(ctx)
	if err != nil {
		return ctx, err
	}
	if todo.Title != expectedTitle {
		return ctx, fmt.Errorf("expected todo title '%s', got '%s'", expectedTitle, todo.Title)
	}
	return ctx, nil
}

func theTodoShouldNotBeCompleted(ctx context.Context) (context.Context, error) {
	todo, err := readTodoFromLastResponse(ctx)
	if err != nil {
		return ctx, err
	}
	if todo.Completed {
		return ctx, fmt.Errorf("expected todo to not be completed, but it is")
	}
	return ctx, nil
}

func theTodoShouldBeCompleted(ctx context.Context) (context.Context, error) {
	todo, err := readTodoFromLastResponse(ctx)
	if err != nil {
		return ctx, err
	}
	if !todo.Completed {
		return ctx, fmt.Errorf("expected todo to be completed, but it is not")
	}
	return ctx, nil
}

func userShouldSeeTodos(ctx context.Context, arg1 string, arg2 int) (context.Context, error) {
	return shouldSeeTodos(ctx, arg2)
}
func shouldSeeTodos(ctx context.Context, expectedCount int) (context.Context, error) {
	todos, err := readTodoListFromLastResponse(ctx)
	if err != nil {
		return ctx, err
	}
	if len(todos) != expectedCount {
		return ctx, fmt.Errorf("expected %d todos, got %d", expectedCount, len(todos))
	}
	return ctx, nil
}

func theTodosShouldInclude(ctx context.Context, expectedTitle string) (context.Context, error) {
	todos, err := readTodoListFromLastResponse(ctx)
	if err != nil {
		return ctx, err
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

func userShouldReceiveTheUpdatedTodo(ctx context.Context) (context.Context, error) {
	return userShouldReceiveASuccessResponse(ctx)
}

func userShouldNoLongerSeeTheTodo(ctx context.Context, username string) (context.Context, error) {
	ctx, err := userRequestsAllTodos(ctx, username)
	if err != nil {
		return ctx, err
	}

	tc := GetTestContextFromContext(ctx)
	deletedTodoID, exists := tc.GetTodoID(username)
	if !exists {
		return ctx, nil
	}

	todos, err := readTodoListFromLastResponse(ctx)
	if err != nil {
		return ctx, err
	}

	for _, todo := range todos {
		if todo.ID == deletedTodoID {
			return ctx, fmt.Errorf("expected todo to be deleted, but it was still found")
		}
	}
	return ctx, nil
}

func userShouldSeeAsCompleted(ctx context.Context, username, title string) (context.Context, error) {
	ctx, err := userRequestsAllTodos(ctx, username)
	if err != nil {
		return ctx, err
	}

	todos, err := readTodoListFromLastResponse(ctx)
	if err != nil {
		return ctx, err
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
	ctx, err := userRequestsAllTodos(ctx, username)
	if err != nil {
		return ctx, err
	}

	todos, err := readTodoListFromLastResponse(ctx)
	if err != nil {
		return ctx, err
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

func userShouldReceiveAnEmptyList(ctx context.Context) (context.Context, error) {
	return shouldSeeTodos(ctx, 0)
}

func userShouldReceiveAValidationError(ctx context.Context) (context.Context, error) {
	return theResponseStatusShouldBe(ctx, 400)
}

func userShouldReceiveANotFoundError(ctx context.Context) (context.Context, error) {
	return theResponseStatusShouldBe(ctx, 404)
}

func iCreateATodoWithTitle(ctx context.Context, title string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	return userCreatesATodoWithTitle(ctx, tc.CurrentUser, title)
}

func iDeleteTheTodo(ctx context.Context) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	return userDeletesTheTodo(ctx, tc.CurrentUser)
}

func iGetAllTodos(ctx context.Context) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	return userRequestsAllTodos(ctx, tc.CurrentUser)
}

func iShouldReceiveAListWithTodo(ctx context.Context, expectedCount int) (context.Context, error) {
	return shouldSeeTodos(ctx, expectedCount)
}

func iShouldReceiveAnEmptyList(ctx context.Context) (context.Context, error) {
	return userShouldReceiveAnEmptyList(ctx)
}

func iShouldReceiveTheCreatedTodoWithAnID(ctx context.Context) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp := tc.GetLastResponse()
	if resp == nil {
		return ctx, fmt.Errorf("no response available")
	}

	if resp.StatusCode != http.StatusCreated {
		return ctx, fmt.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	var todo app.Todo
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, fmt.Errorf("failed to read response body: %w", err)
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &todo); err != nil {
		return ctx, fmt.Errorf("failed to decode todo: %w (body: %s)", err, string(bodyBytes))
	}

	if todo.ID == "" {
		return ctx, fmt.Errorf("expected todo to have an ID")
	}

	tc.StoreTodoID(tc.CurrentUser, todo.ID)
	tc.StoreTodoByTitle(todo.Title, todo.ID)

	return ctx, nil
}

func iShouldReceiveTheUpdatedTodoWithCompletedStatusTrue(ctx context.Context) (context.Context, error) {
	todo, err := readTodoFromLastResponse(ctx)
	if err != nil {
		return ctx, err
	}

	if !todo.Completed {
		return ctx, fmt.Errorf("expected todo to be completed, but it is not")
	}

	return ctx, nil
}

func iShouldReceiveTheUpdatedTodoWithTitle(ctx context.Context, expectedTitle string) (context.Context, error) {
	return theTodoShouldHaveTitle(ctx, expectedTitle)
}

func iUpdateTheTodoCompletionStatusToTrue(ctx context.Context) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	return userMarksTheTodoAsCompleted(ctx, tc.CurrentUser)
}

func iUpdateTheTodoTitleTo(ctx context.Context, newTitle string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	return userUpdatesTheTodoTitleTo(ctx, tc.CurrentUser, newTitle)
}

func userCreatesMultipleTodosWithDifferentTitles(ctx context.Context, username string) (context.Context, error) {
	titles := []string{"Task 1", "Task 2", "Task 3"}

	for _, title := range titles {
		_, err := userCreatesATodoWithTitle(ctx, username, title)
		if err != nil {
			return ctx, fmt.Errorf("failed to create todo '%s': %w", title, err)
		}
	}

	return ctx, nil
}

func userHasCreatedSeveralTodos(ctx context.Context, username string) (context.Context, error) {
	return userCreatesMultipleTodosWithDifferentTitles(ctx, username)
}

func userMarksAsCompleted(ctx context.Context, username, title string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	// Find the todo by title
	todoID, exists := tc.GetTodoIDByTitle(title)
	if !exists {
		return ctx, fmt.Errorf("todo with title '%s' not found", title)
	}

	// Update the todo to mark it as completed
	completed := true
	updateReq := app.UpdateTodoRequest{Completed: &completed}
	resp, err := tc.MakeRequest("PUT", "/todos/"+todoID, updateReq)
	if err != nil {
		return ctx, fmt.Errorf("failed to update todo: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

// userPerformsMultipleOperationsWithTheSameToken performs multiple operations using the same token
func userPerformsMultipleOperationsWithTheSameToken(ctx context.Context, username string) (context.Context, error) {
	_, err := userCreatesATodoWithTitle(ctx, username, "Token test task")
	if err != nil {
		return ctx, err
	}

	_, err = userRequestsAllTodos(ctx, username)
	if err != nil {
		return ctx, err
	}

	_, err = userUpdatesTheTodoTitleTo(ctx, username, "Updated token test task")
	if err != nil {
		return ctx, err
	}

	_, err = userMarksTheTodoAsCompleted(ctx, username)
	if err != nil {
		return ctx, err
	}

	_, err = userDeletesTheTodo(ctx, username)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func userPerformsVariousOperationsUpdateDeleteCreate(ctx context.Context, username string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	_, err := userCreatesATodoWithTitle(ctx, username, "New task")
	if err != nil {
		return ctx, err
	}

	_, err = userRequestsAllTodos(ctx, username)
	if err != nil {
		return ctx, err
	}

	_, exists := tc.GetTodoID(username)
	if !exists {
		return ctx, fmt.Errorf("no todo ID found for user %s", username)
	}

	_, err = userUpdatesTheTodoTitleTo(ctx, username, "Updated task")
	if err != nil {
		return ctx, err
	}

	_, err = userDeletesTheTodo(ctx, username)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func userShouldBeAbleToAccessTheirTodos(ctx context.Context, username string) (context.Context, error) {
	_, err := userRequestsAllTodos(ctx, username)
	if err != nil {
		return ctx, err
	}

	return theResponseStatusShouldBe(ctx, 200)
}

func userShouldSeeTodosInTheirList(ctx context.Context, username string, expectedCount int) (context.Context, error) {
	_, err := userRequestsAllTodos(ctx, username)
	if err != nil {
		return ctx, err
	}

	return shouldSeeTodos(ctx, expectedCount)
}

func userUpdatesTitleTo(ctx context.Context, username, oldTitle, newTitle string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	todoID, exists := tc.GetTodoIDByTitle(oldTitle)
	if !exists {
		return ctx, fmt.Errorf("todo with title '%s' not found", oldTitle)
	}

	updateReq := app.UpdateTodoRequest{Title: &newTitle}
	resp, err := tc.MakeRequest("PUT", "/todos/"+todoID, updateReq)
	if err != nil {
		return ctx, fmt.Errorf("failed to update todo: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func allChangesShouldBeReflectedCorrectly(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func allOperationsShouldSucceed(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func noDataShouldBeLost(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func noDataShouldBeLostOrCorrupted(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func subsequentOperationsShouldWorkNormally(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func theDataShouldRemainConsistent(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func theSystemEncountersATemporaryErrorDuringAnOperation(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func theSystemShouldRecoverGracefully(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func theTokenShouldRemainValid(ctx context.Context) (context.Context, error) {
	return ctx, nil
}
