Feature: User Authorization and Data Isolation
  In order to protect user data
  As a system
  I want to ensure users can only access their own todos

  Background:
    Given the secret key "test-secret" is set up
    And a user named "alice" with password "password123" is registered
    And a user named "bob" with password "password456" is registered
    And user "alice" logs in with password "password123" successfully
    And user "bob" logs in with password "password456" successfully

  Scenario: Users can only see their own todos
    Given user "alice" has created a todo with title "Alice's Task"
    And user "bob" has created a todo with title "Bob's Task"
    When user "alice" gets all todos
    Then user "alice" should only see "Alice's Task"
    And user "alice" should not see "Bob's Task"

  Scenario: Users cannot access each other's todos
    Given user "alice" has created a todo with title "Alice's Private Task"
    And user "bob" has created a todo with title "Bob's Private Task"
    When user "alice" tries to get user "bob"'s todo by ID
    Then user "alice" should receive an error message "Todo not found"
    And the response status should be 404

  Scenario: Users cannot update each other's todos
    Given user "alice" has created a todo with title "Alice's Task"
    And user "bob" has created a todo with title "Bob's Task"
    When user "bob" tries to update user "alice"'s todo
    Then user "bob" should receive an error message "Todo not found"
    And the response status should be 404
  Scenario: Users cannot delete each other's todos
    Given user "alice" has created a todo with title "Alice's Task"
    And user "bob" has created a todo with title "Bob's Task"
    When user "bob" tries to delete user "alice"'s todo
    Then user "bob" should receive an error message "Todo not found"
    And the response status should be 404

  Scenario: Invalid token access
    When I try to access todos with an invalid token
    Then I should receive an error message "Invalid token"
    And the response status should be 401

  Scenario: Expired token access
    When I try to access todos with an expired token
    Then I should receive an error message "Invalid token"
    And the response status should be 401

  Scenario: Missing authorization header
    When I try to access todos without an authorization header
    Then I should receive an error message "Authorization header required"
    And the response status should be 401

  Scenario: Malformed authorization header
    When I try to access todos with malformed authorization header
    Then I should receive an error message "Invalid token"
    And the response status should be 401