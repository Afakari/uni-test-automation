Feature: End-to-End Integration Tests
  In order to verify complete user workflows
  As a system
  I want to ensure all components work together correctly

  Background:
    Given the secret key "test-secret" is set up

  Scenario: Complete user journey - registration to todo management
    When I register with username "newuser" and password "mypassword"
    Then I should receive a success message
    And the response status should be 201

    When I login with username "newuser" and password "mypassword"
    Then I should receive a valid JWT token
    And the response status should be 200

    When I create a todo with title "My first task"
    Then I should receive the created todo with an ID
    And the response status should be 201

    When I get all todos
    Then I should receive a list with 1 todo
    And the todo should have title "My first task"
    And the response status should be 200

    When I update the todo title to "Updated first task"
    Then I should receive the updated todo with title "Updated first task"
    And the response status should be 200

    When I update the todo completion status to true
    Then I should receive the updated todo with completed status true
    And the response status should be 200

    When I delete the todo
    Then I should receive a success message "Todo deleted"
    And the response status should be 200

    When I get all todos
    Then I should receive an empty list
    And the response status should be 200

  Scenario: Multiple users with separate data
    Given a user named "alice" with password "alice123" is registered
    And a user named "bob" with password "bob456" is registered
    And user "alice" logs in with password "alice123" successfully
    And user "bob" logs in with password "bob456" successfully

    When user "alice" creates a todo with title "Alice's shopping list"
    And user "bob" creates a todo with title "Bob's work tasks"
    And user "alice" creates a todo with title "Alice's personal goals"
    And user "bob" creates a todo with title "Bob's weekend plans"

    Then user "alice" should see 2 todos in their list
    And user "alice" should not see "Bob's work tasks"
    And user "bob" should see 2 todos in their list
    And user "bob" should not see "Alice's shopping list"

  Scenario: Error recovery and system resilience
    Given a user named "alice" with password "password123" is registered
    And user "alice" logs in with password "password123" successfully
    And user "alice" has created several todos

    When the system encounters a temporary error during an operation
    Then the system should recover gracefully
    And subsequent operations should work normally
    And no data should be lost