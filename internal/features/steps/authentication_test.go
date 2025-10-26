package steps

import (
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
	"todoapp/internal/app"
	"todoapp/internal/features/support"
)

func TestUserAuthentication(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeUserAuthScenario,
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

func InitializeUserAuthScenario(ctx *godog.ScenarioContext) {
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

	ctx.Step(`^the secret key "([^"]*)" is set up$`, theSecretKeyIsSetUp)
	ctx.Step(`^a user named "([^"]*)" with password "([^"]*)" is already registered$`, aUserNamedWithPasswordIsRegistered)

	ctx.Step(`^I register with username "([^"]*)" and password "([^"]*)"$`, iRegisterWithUsernameAndPassword)
	ctx.Step(`^I login with username "([^"]*)" and password "([^"]*)"$`, iLoginWithUsernameAndPassword)
	ctx.Step(`^I send invalid JSON to the register endpoint$`, iSendInvalidJSONToTheRegisterEndpoint)
	ctx.Step(`^I send invalid JSON to the login endpoint$`, iSendInvalidJSONToTheLoginEndpoint)
	ctx.Step(`^a user named "([^"]*)" with password "([^"]*)" is registered$`, aUserNamedWithPasswordIsRegistered)

	ctx.Step(`^I should receive a success message$`, iShouldReceiveASuccessMessage)
	ctx.Step(`^I should receive a valid JWT token$`, iShouldReceiveAValidJwtToken)
	ctx.Step(`^I should receive an error message "([^"]*)"$`, iShouldReceiveAnErrorMessage)
	ctx.Step(`^the response status should be (\d+)$`, theResponseStatusShouldBe)
}

func iRegisterWithUsernameAndPassword(ctx context.Context, username, password string) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
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
	tc := support.GetTestContextFromContext(ctx)
	if tc == nil {
		return ctx, fmt.Errorf("test context not found")
	}

	creds := app.Credentials{Username: username, Password: password}
	resp, err := tc.MakeRequest("POST", "/login", creds)
	if err != nil {
		return ctx, fmt.Errorf("failed to login user: %w", err)
	}

	tc.SetLastResponse(resp)
	return ctx, nil
}

func iSendInvalidJSONToTheRegisterEndpoint(ctx context.Context) (context.Context, error) {
	tc := support.GetTestContextFromContext(ctx)
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
	tc := support.GetTestContextFromContext(ctx)
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

func iShouldReceiveASuccessMessage(ctx context.Context) (context.Context, error) {
	return userShouldReceiveASuccessMessage(ctx, "")
}

func iShouldReceiveAValidJwtToken(ctx context.Context) (context.Context, error) {
	return userShouldReceiveAValidJwtToken(ctx, "")
}

func iShouldReceiveAnErrorMessage(ctx context.Context, expectedError string) (context.Context, error) {
	return userShouldReceiveAnErrorMessage(ctx, "", expectedError)
}
