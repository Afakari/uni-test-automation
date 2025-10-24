Feature: User Authentication
  In order to access the todo management system
  As a user
  I want to be able to register and login securely

  Background:
    Given the secret key "test-secret" is set up

  Scenario: Successful user registration
    When I register with username "alice" and password "password123"
    Then I should receive a success message
    And the response status should be 201

  Scenario: Registration with missing username
    When I register with username "" and password "password123"
    Then I should receive an error message "Username and password required"
    And the response status should be 400

  Scenario: Registration with missing password
    When I register with username "alice" and password ""
    Then I should receive an error message "Username and password required"
    And the response status should be 400

  Scenario: Registration with invalid JSON
    When I send invalid JSON to the register endpoint
    Then I should receive an error message "Invalid request"
    And the response status should be 400

  Scenario: Duplicate user registration
    Given a user named "alice" with password "password123" is already registered
    When I register with username "alice" and password "password123"
    Then I should receive an error message "User already exists"
    And the response status should be 409

  Scenario: Successful user login
    Given a user named "alice" with password "password123" is registered
    When I login with username "alice" and password "password123"
    Then I should receive a valid JWT token
    And the response status should be 200

  Scenario: Login with invalid username
    Given a user named "alice" with password "password123" is registered
    When I login with username "bob" and password "password123"
    Then I should receive an error message "Invalid credentials"
    And the response status should be 401

  Scenario: Login with invalid password
    Given a user named "alice" with password "password123" is registered
    When I login with username "alice" and password "wrongpassword"
    Then I should receive an error message "Invalid credentials"
    And the response status should be 401

  Scenario: Login with missing credentials
    When I login with username "" and password ""
    Then I should receive an error message "Username and password required"
    And the response status should be 400

  Scenario: Login with invalid JSON
    When I send invalid JSON to the login endpoint
    Then I should receive an error message "Invalid request"
    And the response status should be 400