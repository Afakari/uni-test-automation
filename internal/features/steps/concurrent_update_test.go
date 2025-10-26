package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"todoapp/internal/app"
)

var concurrentWG sync.WaitGroup

func aNewTaskTitledIsCreated(ctx context.Context, title string) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	reqBody := map[string]string{"title": title}

	resp, err := tc.MakeRequest("POST", "/todos", reqBody)
	if err != nil {
		return ctx, err
	}

	if resp.StatusCode != http.StatusCreated {
		return ctx, fmt.Errorf("expected status code 201, got %v", resp.StatusCode)
	}

	var created app.Todo
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, err
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &created); err != nil {
		return ctx, err
	}

	tc.CurrentTodoID = created.ID
	tc.OriginalTitle = created.Title
	return ctx, nil
}

func theFirstPersonTriesToChangeTheTask(ctx context.Context, title string) (context.Context, error) {
	const loops = 200
	concurrentWG.Add(1)
	go runConcurrentAppends(ctx, title, loops)
	return ctx, nil
}

func theSecondPersonTriesToChangeTheTask(ctx context.Context, title string) (context.Context, error) {
	const loops = 200
	concurrentWG.Add(1)
	go runConcurrentAppends(ctx, title, loops)
	concurrentWG.Wait()
	return ctx, nil
}

func bothPeopleShouldReceiveConfirmationThatTheirChangesWereSaved(ctx context.Context) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	const totalExpectedUpdates = 400
	tc.concMutex.Lock()
	defer tc.concMutex.Unlock()

	if tc.success != totalExpectedUpdates {
		return ctx, fmt.Errorf("expected %d successful updates, got %d. Errors: %d", totalExpectedUpdates, tc.success, len(tc.errs))
	}
	if len(tc.errs) > 0 {
		return ctx, fmt.Errorf("expected 0 errors, got %d concurrent failures", len(tc.errs))
	}
	return ctx, nil
}

func theFinalTitleOfTheTaskWillBeARandomMixOfTheTwoVersions(ctx context.Context) (context.Context, error) {
	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	resp, err := tc.MakeRequest("GET", "/todos/"+tc.CurrentTodoID, nil)
	if err != nil {
		return ctx, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ctx, fmt.Errorf("final GET failed with status %d", resp.StatusCode)
	}

	var todo app.Todo
	if err := json.NewDecoder(resp.Body).Decode(&todo); err != nil {
		return ctx, err
	}

	finalTitle := todo.Title
	countA := strings.Count(finalTitle, "A")
	countB := strings.Count(finalTitle, "B")
	totalAppends := countA + countB
	originalLen := len(tc.OriginalTitle)
	finalLen := len(finalTitle)

	fmt.Printf("\n--- Test Result Explanation ---\n")
	fmt.Printf("Initial Task Title: '%s' (Length: %d)\n", tc.OriginalTitle, originalLen)
	fmt.Printf("Final Task Title:   '%s' (Length: %d)\n", finalTitle, finalLen)
	fmt.Printf("\nCompeting Appends:\n")
	fmt.Printf("  - Person 1: Tried to append 'A' 200 times\n")
	fmt.Printf("  - Person 2: Tried to append 'B' 200 times\n")
	fmt.Printf("Total Expected Appends: 400\n")
	fmt.Printf("\nActual Result:\n")
	fmt.Printf("  - 'A's in final title: %d\n", countA)
	fmt.Printf("  - 'B's in final title: %d\n", countB)
	fmt.Printf("  - Total Appends Captured: %d\n", totalAppends)

	if totalAppends == 400 {
		return ctx, fmt.Errorf("TEST FAILED: NO RACE CONDITION DETECTED. All 400 appends were saved")
	}
	if totalAppends == 0 {
		return ctx, fmt.Errorf("TEST FAILED: Zero appends were saved. Something is wrong")
	}

	fmt.Printf("\nConclusion (Race Condition Demonstrated):\n")
	fmt.Printf("The system successfully processed all 400 update requests (all returned 200 OK).\n")
	fmt.Printf("However, because both users were reading the *same* old data, modifying it, and writing it back,\n")
	fmt.Printf("they overwrote each other's changes. This is a classic 'Lost Update' race condition.\n")
	fmt.Printf("Only %d out of 400 changes were actually saved to the database.\n", totalAppends)
	fmt.Printf("-------------------------------\n")

	return ctx, nil
}

func runConcurrentAppends(ctx context.Context, charToAppend string, updates int) {
	defer concurrentWG.Done()

	tc := GetTestContextFromContext(ctx)
	if tc == nil {
		fmt.Println("Error: TestContext not found in goroutine")
		return
	}

	for i := 0; i < updates; i++ {
		getResp, err := tc.MakeRequest("GET", "/todos/"+tc.CurrentTodoID, nil)
		if err != nil {
			tc.concMutex.Lock()
			tc.errs = append(tc.errs, fmt.Errorf("GET request error: %w", err))
			tc.concMutex.Unlock()
			continue
		}

		var currentTodo app.Todo
		if err := json.NewDecoder(getResp.Body).Decode(&currentTodo); err != nil {
			getResp.Body.Close()
			tc.concMutex.Lock()
			tc.errs = append(tc.errs, fmt.Errorf("GET decode error: %w", err))
			tc.concMutex.Unlock()
			continue
		}
		getResp.Body.Close()

		if getResp.StatusCode != http.StatusOK {
			tc.concMutex.Lock()
			tc.errs = append(tc.errs, fmt.Errorf("GET failed with status %d", getResp.StatusCode))
			tc.concMutex.Unlock()
			continue
		}

		newTitle := currentTodo.Title + charToAppend
		req := app.UpdateTodoRequest{Title: &newTitle}

		putResp, err := tc.MakeRequest("PUT", "/todos/"+tc.CurrentTodoID, req)
		if err != nil {
			tc.concMutex.Lock()
			tc.errs = append(tc.errs, fmt.Errorf("PUT request error: %w", err))
			tc.concMutex.Unlock()
			continue
		}

		tc.concMutex.Lock()
		if putResp.StatusCode == http.StatusOK {
			tc.success++
		} else {
			tc.errs = append(tc.errs, fmt.Errorf("PUT update failed with status %d", putResp.StatusCode))
		}
		tc.concMutex.Unlock()
		putResp.Body.Close()
	}
}
