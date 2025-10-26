Feature: Todo Management
  In order to manage my tasks effectively
  As a user
  I want to create, read, update, and delete todo items

  Background:
    Given the secret key "test-secret" is set up
    And a user named "alice" with password "password123" is registered
    And user "alice" logs in with password "password123" successfully

  Scenario: User can create a new todo
    When user "alice" creates a todo with title "Buy groceries"
    Then user "alice" should receive a created response
    And the todo should have title "Buy groceries"
    And the todo should not be completed

  Scenario: User can retrieve all their todos
    Given user "alice" has created a todo with title "Task 1"
    And user "alice" has created a todo with title "Task 2"
    When user "alice" requests all todos
    Then user "alice" should see 2 todos
    And the todos should include "Task 1"
    And the todos should include "Task 2"

  Scenario: User can retrieve a specific todo
    Given user "alice" has created a todo with title "Specific task"
    When user "alice" requests the todo by ID
    Then user "alice" should receive the todo with title "Specific task"

  Scenario: User can update a todo title
    Given user "alice" has created a todo with title "Original title"
    When user "alice" updates the todo title to "Updated title"
    Then user "alice" should receive the updated todo
    And the todo should have title "Updated title"

  Scenario: User can mark a todo as completed
    Given user "alice" has created a todo with title "Task to complete"
    When user "alice" marks the todo as completed
    Then user "alice" should receive the updated todo
    And the todo should be completed

  Scenario: User can delete a todo
    Given user "alice" has created a todo with title "Task to delete"
    When user "alice" deletes the todo
    Then user "alice" should receive a success message
    And user "alice" should no longer see the todo

  Scenario: User cannot access another user's todos
    Given a user named "bob" with password "password456" is registered
    And user "bob" logs in with password "password456" successfully
    And user "bob" has created a todo with title "Bob's private task"
    And user "alice" logs in with password "password123" successfully
    When user "alice" requests all todos
    Then user "alice" should not see "Bob's private task"

  Scenario: User cannot modify another user's todo
    Given a user named "bob" with password "password456" is registered
    And user "bob" logs in with password "password456" successfully
    And user "bob" has created a todo with title "Bob's task"
    And user "alice" logs in with password "password123" successfully
    When user "alice" tries to update Bob's todo
    Then user "alice" should receive a not found error

  Scenario: User can manage multiple todos independently
    Given user "alice" has created a todo with title "Work task"
    And user "alice" has created a todo with title "Personal task"
    When user "alice" marks "Work task" as completed
    And user "alice" updates "Personal task" title to "Updated personal task"
    Then user "alice" should see "Work task" as completed
    And user "alice" should see "Updated personal task" as not completed

  Scenario: System handles empty todo list gracefully
    When user "alice" requests all todos
    Then user "alice" should receive an empty list
    And the response status should be 200

  Scenario: System validates todo creation requirements
    When user "alice" tries to create a todo with empty title
    Then user "alice" should receive a validation error
    And the response status should be 400

  Scenario: System handles non-existent todo operations
    When user "alice" tries to update a non-existent todo
    Then user "alice" should receive a not found error
    And the response status should be 404

    When user "alice" tries to delete a non-existent todo
    Then user "alice" should receive a not found error
    And the response status should be 404
