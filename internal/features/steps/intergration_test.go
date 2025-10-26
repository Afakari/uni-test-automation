package steps

import (
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
	"todoapp/internal/features/support"
)

func TestIntegration(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeIntegrationTest,
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

func InitializeIntegrationTest(ctx *godog.ScenarioContext) {
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

	// Use all the common step definitions
	ctx.Step(`^the secret key "([^"]*)" is set up$`, theSecretKeyIsSetUp)
	ctx.Step(`^a user named "([^"]*)" with password "([^"]*)" is already registered$`, aUserNamedWithPasswordIsRegistered)
	ctx.Step(`^a user named "([^"]*)" with password "([^"]*)" is registered$`, aUserNamedWithPasswordIsRegistered)

	ctx.Step(`^I register with username "([^"]*)" and password "([^"]*)"$`, iRegisterWithUsernameAndPassword)
	ctx.Step(`^I login with username "([^"]*)" and password "([^"]*)"$`, iLoginWithUsernameAndPassword)
	ctx.Step(`^I send invalid JSON to the register endpoint$`, iSendInvalidJSONToTheRegisterEndpoint)
	ctx.Step(`^I send invalid JSON to the login endpoint$`, iSendInvalidJSONToTheLoginEndpoint)

	ctx.Step(`^I should receive a success message$`, iShouldReceiveASuccessMessage)
	ctx.Step(`^I should receive a valid JWT token$`, iShouldReceiveAValidJwtToken)
	ctx.Step(`^I should receive an error message "([^"]*)"$`, iShouldReceiveAnErrorMessage)
	ctx.Step(`^the response status should be (\d+)$`, theResponseStatusShouldBe)
}
