package steps

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"todoapp/internal/app"
)
import "context"
import "github.com/cucumber/godog"

var tf *todoFeature

func TestConcurrent_update(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeconcurrentUpdatescenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../concurrent_update.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeconcurrentUpdatescenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		tf = &todoFeature{}
		ctx, _ = setupServer(ctx, tf)
		return ctx, nil
	})
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		ctx, err = closeSever(ctx, tf)
		return ctx, err
	})
	ctx.Step(`^the secret key "([^"]*)" is set up$`, isJwtSecretSet)
	ctx.Step(`^a user named "([^"]*)" with password "([^"]*)" is registered$`, aUserNamedAliceWithPasswordPassIsRegistered)
	ctx.Step(`^user "([^"]*)" logs in with password "([^"]*)" successfully$`, userAliceLogsInWithPasswordPassSuccessfully)
	ctx.Step(`^a new task titled "([^"]*)" is created$`, aNewTaskTitledConcurrentTaskIsCreated)
	ctx.Step(`^the first person tries to change the task's title to "([^"]*)" many times$`, theFirstPersonTriesToChangeTheTask)
	ctx.Step(`^the second person tries to change the task's title to "([^"]*)" many times$`, theSecondPersonTriesToChangeTheTask)
	ctx.Step(`^both people should receive confirmation that their changes were saved$`, bothPeopleShouldReceiveConfirmationThatTheirChangesWereSaved)
	ctx.Step(`^the final title of the task will be a random mix of the two versions$`, theFinalTitleOfTheTaskWillBeARandomMixOfTheTwoVersions)
}
func isJwtSecretSet(ctx context.Context, secret string) (context.Context, error) {
	if err := os.Setenv("JWT_SECRET", secret); err != nil {
		return ctx, err
	}

	app.Init(secret)
	app.Users = sync.Map{}
	app.Todos = sync.Map{}

	return ctx, nil
}
func aUserNamedAliceWithPasswordPassIsRegistered(ctx context.Context, username string, password string) (context.Context, error) {
	body := app.Credentials{Username: username, Password: password}

	resp, err := tf.makeRequest("POST", "/register", body)
	if err != nil {
		return ctx, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return ctx, fmt.Errorf("expected status code 201, got %v", resp.StatusCode)
	}

	return ctx, nil
}
func userAliceLogsInWithPasswordPassSuccessfully(ctx context.Context, username string, password string) (context.Context, error) {
	body := app.Credentials{Username: username, Password: password}
	resp, err := tf.makeRequest("POST", "/login", body)
	if err != nil {
		return ctx, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return ctx, fmt.Errorf("expected status code 200, got %v", resp.StatusCode)
	}

	var resBody map[string]string

	if err := json.NewDecoder(resp.Body).Decode(&resBody); err != nil {
		return ctx, err
	}

	tf.token = resBody["token"]
	if tf.token == "" {
		return ctx, fmt.Errorf("login failed, no token returned")
	}

	return ctx, nil
}
func aNewTaskTitledConcurrentTaskIsCreated(ctx context.Context, title string) (context.Context, error) {
	reqBody := map[string]string{"title": title}

	// the helper function automatically sets the token if available
	resp, err := tf.makeRequest("POST", "/todos", reqBody)

	if err != nil {
		return ctx, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return ctx, fmt.Errorf("expected status code 201, got %v", resp.StatusCode)
	}
	var created app.Todo
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return ctx, err
	}

	tf.todoID = created.ID
	tf.originalTitle = created.Title
	return ctx, nil
}

func runConcurrentAppends(tf *todoFeature, charToAppend string, updates int) {
	defer concurrentWG.Done()

	for i := 0; i < updates; i++ {
		tf.recordAppendAttempt(charToAppend)
	}
}

func theFirstPersonTriesToChangeTheTask(ctx context.Context, title string) (context.Context, error) {
	const loops = 200
	concurrentWG.Add(1)
	go runConcurrentAppends(tf, title, loops)
	return ctx, nil
}

func theSecondPersonTriesToChangeTheTask(ctx context.Context, title string) (context.Context, error) {
	const loops = 200
	concurrentWG.Add(1)
	go runConcurrentAppends(tf, title, loops)
	concurrentWG.Wait()
	return ctx, nil
}
func bothPeopleShouldReceiveConfirmationThatTheirChangesWereSaved(ctx context.Context) (context.Context, error) {
	const totalExpectedUpdates = 400
	stateMutex.Lock()
	defer stateMutex.Unlock()

	if tf.success != totalExpectedUpdates {
		return ctx, fmt.Errorf("expected %d successful updates, got %d. Errors: %d", totalExpectedUpdates, tf.success, len(tf.errs))
	}
	if len(tf.errs) > 0 {
		return ctx, fmt.Errorf("expected 0 errors, got %d concurrent failures", len(tf.errs))
	}
	return ctx, nil
}

func theFinalTitleOfTheTaskWillBeARandomMixOfTheTwoVersions(ctx context.Context) (context.Context, error) {
	resp, err := tf.makeRequest("GET", "/todos/"+tf.todoID, nil)
	if err != nil {
		return ctx, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

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
	originalLen := len(tf.originalTitle)
	finalLen := len(finalTitle)

	fmt.Printf("\n--- Test Result Explanation ---\n")
	fmt.Printf("Initial Task Title: '%s' (Length: %d)\n", tf.originalTitle, originalLen)
	fmt.Printf("Final Task Title:   '%s' (Length: %d)\n", finalTitle, finalLen)
	fmt.Printf("\nCompeting Appends:\n")
	fmt.Printf("  - Person 1: Tried to append 'A' 200 times\n")
	fmt.Printf("  - Person 2: Tried to append 'B' 200 times\n")
	fmt.Printf("Total Expected Appends: 400\n")
	fmt.Printf("\nActual Result:\n")
	fmt.Printf("  - 'A's in final title: %d\n", countA)
	fmt.Printf("  - 'B's in final title: %d\n", countB)
	fmt.Printf("  - **Total Appends Captured: %d**\n", totalAppends)

	// The proof:
	if totalAppends == 400 {
		return ctx, fmt.Errorf("TEST FAILED: NO RACE CONDITION DETECTED. All 400 appends were saved.")
	}
	if totalAppends == 0 {
		return ctx, fmt.Errorf("TEST FAILED: Zero appends were saved. Something is wrong.")
	}

	fmt.Printf("\n**Conclusion (Race Condition Demonstrated):**\n")
	fmt.Printf("The system successfully processed all 400 update requests (all returned 200 OK).\n")
	fmt.Printf("However, because both users were reading the *same* old data, modifying it, and writing it back,\n")
	fmt.Printf("they overwrote each other's changes. This is a classic 'Lost Update' race condition.\n")
	fmt.Printf("Only %d out of 400 changes were actually saved to the database.\n", totalAppends)
	fmt.Printf("-------------------------------\n")

	return ctx, nil
}

func setupServer(ctx context.Context, tf *todoFeature) (context.Context, error) {
	router := app.SetupRouter()

	tf.server = httptest.NewServer(router)
	tf.baseURL = tf.server.URL
	tf.client = tf.server.Client()
	tf.errs = nil
	tf.success = 0
	return ctx, nil
}

func closeSever(ctx context.Context, tf *todoFeature) (context.Context, error) {
	if tf.server != nil {
		tf.server.Close()
	}
	return ctx, nil
}
