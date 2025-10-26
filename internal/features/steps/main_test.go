package steps

import (
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenarios,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{".."},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenarios(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		tc := NewTestContext()
		if err := tc.SetupServer(); err != nil {
			return ctx, fmt.Errorf("failed to setup test server: %w", err)
		}
		return SetTestContextInContext(ctx, tc), nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		tc := GetTestContextFromContext(ctx)
		if tc != nil {
			tc.CloseServer()
		}
		return ctx, nil
	})

	ctx.Step(`^the secret key "([^"]*)" is set up$`, theSecretKeyIsSetUp)
	ctx.Step(`^a user named "([^"]*)" with password "([^"]*)" is already registered$`, aUserNamedWithPasswordIsRegistered)
	ctx.Step(`^a user named "([^"]*)" with password "([^"]*)" is registered$`, aUserNamedWithPasswordIsRegistered)
	ctx.Step(`^user "([^"]*)" logs in with password "([^"]*)" successfully$`, userLogsInWithPasswordSuccessfully)
	ctx.Step(`^the response status should be (\d+)$`, theResponseStatusShouldBe)
	ctx.Step(`^user "([^"]*)" should receive a success message$`, userShouldReceiveASuccessMessage)
	ctx.Step(`^user "([^"]*)" should receive a valid JWT token$`, userShouldReceiveAValidJwtToken)
	ctx.Step(`^user "([^"]*)" should receive an error message "([^"]*)"$`, userShouldReceiveAnErrorMessage)
	ctx.Step(`^I should receive a success message$`, iShouldReceiveASuccessMessage)
	ctx.Step(`^I should receive a valid JWT token$`, iShouldReceiveAValidJwtToken)
	ctx.Step(`^I should receive an error message "([^"]*)"$`, iShouldReceiveAnErrorMessage)

	ctx.Step(`^I register with username "([^"]*)" and password "([^"]*)"$`, iRegisterWithUsernameAndPassword)
	ctx.Step(`^I login with username "([^"]*)" and password "([^"]*)"$`, iLoginWithUsernameAndPassword)
	ctx.Step(`^I send invalid JSON to the register endpoint$`, iSendInvalidJSONToTheRegisterEndpoint)
	ctx.Step(`^I send invalid JSON to the login endpoint$`, iSendInvalidJSONToTheLoginEndpoint)

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

	ctx.Step(`^user "([^"]*)" creates a todo with title "([^"]*)"$`, userCreatesATodoWithTitle)
	ctx.Step(`^user "([^"]*)" has created a todo with title "([^"]*)"$`, userHasCreatedATodoWithTitle)
	ctx.Step(`^user "([^"]*)" requests all todos$`, userRequestsAllTodos)
	ctx.Step(`^user "([^"]*)" requests the todo by ID$`, userRequestsTheTodoByID)
	ctx.Step(`^user "([^"]*)" updates the todo title to "([^"]*)"$`, userUpdatesTheTodoTitleTo)
	ctx.Step(`^user "([^"]*)" marks the todo as completed$`, userMarksTheTodoAsCompleted)
	ctx.Step(`^user "([^"]*)" deletes the todo$`, userDeletesTheTodo)
	ctx.Step(`^user "([^"]*)" tries to update Bob's todo$`, userTriesToUpdateBobsTodo)
	ctx.Step(`^user "([^"]*)" tries to create a todo with empty title$`, userTriesToCreateATodoWithEmptyTitle)
	ctx.Step(`^user "([^"]*)" tries to update a non-existent todo$`, userTriesToUpdateANonExistentTodo)
	ctx.Step(`^user "([^"]*)" tries to delete a non-existent todo$`, userTriesToDeleteANonExistentTodo)
	ctx.Step(`^user "([^"]*)" should receive a success response$`, userShouldReceiveASuccessResponse)
	ctx.Step(`^user "([^"]*)" should receive a created response$`, userShouldReceiveACreatedResponse)
	ctx.Step(`^the todo should have title "([^"]*)"$`, theTodoShouldHaveTitle)
	ctx.Step(`^the todo should not be completed$`, theTodoShouldNotBeCompleted)
	ctx.Step(`^the todo should be completed$`, theTodoShouldBeCompleted)
	ctx.Step(`^should see "(\d+)" todos$`, shouldSeeTodos)
	ctx.Step(`^the todos should include "([^"]*)"$`, theTodosShouldInclude)
	ctx.Step(`^user "([^"]*)" should receive the todo with title "([^"]*)"$`, userShouldReceiveTheTodoWithTitle)
	ctx.Step(`^user "([^"]*)" should receive the updated todo$`, userShouldReceiveTheUpdatedTodo)
	ctx.Step(`^user "([^"]*)" should no longer see the todo$`, userShouldNoLongerSeeTheTodo)
	ctx.Step(`^user "([^"]*)" should see "([^"]*)" as completed$`, userShouldSeeAsCompleted)
	ctx.Step(`^user "([^"]*)" should see "([^"]*)" as not completed$`, userShouldSeeAsNotCompleted)
	ctx.Step(`^user "([^"]*)" should receive an empty list$`, userShouldReceiveAnEmptyList)
	ctx.Step(`^user "([^"]*)" should receive a validation error$`, userShouldReceiveAValidationError)
	ctx.Step(`^user "([^"]*)" should receive a not found error$`, userShouldReceiveANotFoundError)

	ctx.Step(`^a new task titled "([^"]*)" is created$`, aNewTaskTitledIsCreated)
	ctx.Step(`^the first person tries to change the task's title to "([^"]*)" many times$`, theFirstPersonTriesToChangeTheTask)
	ctx.Step(`^the second person tries to change the task's title to "([^"]*)" many times$`, theSecondPersonTriesToChangeTheTask)
	ctx.Step(`^both people should receive confirmation that their changes were saved$`, bothPeopleShouldReceiveConfirmationThatTheirChangesWereSaved)
	ctx.Step(`^the final title of the task will be a random mix of the two versions$`, theFinalTitleOfTheTaskWillBeARandomMixOfTheTwoVersions)

	ctx.Step(`^all changes should be reflected correctly$`, allChangesShouldBeReflectedCorrectly)
	ctx.Step(`^all operations should succeed$`, allOperationsShouldSucceed)
	ctx.Step(`^I create a todo with title "([^"]*)"$`, iCreateATodoWithTitle)
	ctx.Step(`^I delete the todo$`, iDeleteTheTodo)
	ctx.Step(`^I get all todos$`, iGetAllTodos)
	ctx.Step(`^I should receive a list with (\d+) todo$`, iShouldReceiveAListWithTodo)
	ctx.Step(`^I should receive a success message "([^"]*)"$`, iShouldReceiveASuccessMessage)
	ctx.Step(`^I should receive an empty list$`, iShouldReceiveAnEmptyList)
	ctx.Step(`^I should receive the created todo with an ID$`, iShouldReceiveTheCreatedTodoWithAnID)
	ctx.Step(`^I should receive the updated todo with completed status true$`, iShouldReceiveTheUpdatedTodoWithCompletedStatusTrue)
	ctx.Step(`^I should receive the updated todo with title "([^"]*)"$`, iShouldReceiveTheUpdatedTodoWithTitle)
	ctx.Step(`^I update the todo completion status to true$`, iUpdateTheTodoCompletionStatusToTrue)
	ctx.Step(`^I update the todo title to "([^"]*)"$`, iUpdateTheTodoTitleTo)
	ctx.Step(`^no data should be lost$`, noDataShouldBeLost)
	ctx.Step(`^no data should be lost or corrupted$`, noDataShouldBeLostOrCorrupted)
	ctx.Step(`^subsequent operations should work normally$`, subsequentOperationsShouldWorkNormally)
	ctx.Step(`^the data should remain consistent$`, theDataShouldRemainConsistent)
	ctx.Step(`^the system encounters a temporary error during an operation$`, theSystemEncountersATemporaryErrorDuringAnOperation)
	ctx.Step(`^the system should recover gracefully$`, theSystemShouldRecoverGracefully)
	ctx.Step(`^the token should remain valid$`, theTokenShouldRemainValid)
	ctx.Step(`^user "([^"]*)" creates multiple todos with different titles$`, userCreatesMultipleTodosWithDifferentTitles)
	ctx.Step(`^user "([^"]*)" has created several todos$`, userHasCreatedSeveralTodos)
	ctx.Step(`^user "([^"]*)" marks "([^"]*)" as completed$`, userMarksAsCompleted)
	ctx.Step(`^user "([^"]*)" performs multiple operations with the same token$`, userPerformsMultipleOperationsWithTheSameToken)
	ctx.Step(`^user "([^"]*)" performs various operations \(update, delete, create\)$`, userPerformsVariousOperationsUpdateDeleteCreate)
	ctx.Step(`^user "([^"]*)" should be able to access their todos$`, userShouldBeAbleToAccessTheirTodos)
	ctx.Step(`^user "([^"]*)" should see (\d+) todos in their list$`, userShouldSeeTodosInTheirList)
	ctx.Step(`^user "([^"]*)" updates "([^"]*)" title to "([^"]*)"$`, userUpdatesTitleTo)
	ctx.Step(`^user "([^"]*)" should see (\d+) todos$`, userShouldSeeTodos)

}
