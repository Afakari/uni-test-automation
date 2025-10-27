# Test Automation and BDD with Go and Cucumber

This is an experimental project showcasing test automation and Behavior-Driven Development (BDD) using Go and Cucumber. The project is built for learning purposes and demonstrates the integration of Cucumber with Go using the `godog` library.

## Disclaimer

- The `godog` library (`v0.15.x`) used for running Cucumber tests is currently unstable, and its API is likely to change in the future.
- This project is a one-time experiment, and there are no plans to maintain or update the code to align with future API changes.
- The project structure may not follow recommended production practices, as it was developed based on intuition during my first experience with Cucumber.

## Project Overview

The core of the project is a simple **To-Do List API** with the following features:
- In-memory database
- JWT-based authentication
- Two public APIs:
  - User registration
  - User login (to obtain a JWT token)
- Five protected APIs for CRUD operations on the to-do list

This API serves as the foundation for testing, with both unit tests and BDD tests implemented.

## Project Structure

- Cucumber-related files (Gherkin feature files and step definitions) are located in the `internal/features` directory.
- Step definitions use a custom wrapper around Go's `context` to manage state across stateless functions.
- Helper abstractions are included to improve method reusability. Review these helpers to understand the implementation details.

## Prerequisites

- Go version `1.25.3`
- Git

## Setup and Running the Project

To run the project locally:

1. Clone the repository:
   ```bash
   git clone git@github.com:afakari/uni-test-automation.git
   ```

2. Navigate to the project directory and install dependencies:
   ```bash
   go mod tidy
   ```

3. Install the `godog`:
   ```bash
   go install github.com/cucumber/godog/cmd/godog@0.15.1
   ```

4. Set the JWT secret environment variable:
   ```bash
   export JWT_SECRET=$(openssl rand -base64 32)
   ```

5. Run the application:
   ```bash
   go run .
   ```

This will start the API server using the Gin framework, accessible at `http://localhost:8080`.

## Running Tests

To run the BDD tests, use the provided convenience script:
```bash
./run_bdd_tests.sh
```

This script automatically executes all Cucumber tests located in the `internal/features` directory.

## CI/CD Pipeline

The project includes a GitHub Actions workflow that:
- Runs unit and BDD tests on every commit to the `main` branch.
- Uses a Go-based action environment for testing.
- Publishes the project as a Docker image to `afakari/uni-test-automation` if all tests pass successfully.

## Notes

- This project is an experimental learning exercise and may not reflect best practices for  ANY production use.
- For a deeper understanding of the BDD implementation and my struggles to handle state in an stateless environment, explore the helper functions and step definitions in `internal/features`.

## License

This project is for educational purposes.
