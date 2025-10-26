Feature: Confusion and Data Loss from Simultaneous Updates
  In order to prevent customers from losing information
  As a software developer
  I want to demonstrate that when two people edit the same ToDo at once,
  the system can't decide who wins, leading to unpredictable results.

  Background:
    Given the secret key "test-secret" is set up
    And a user named "Alice" with password "pass" is registered
    And user "Alice" logs in with password "pass" successfully
    And a new task titled "Concurrent Task" is created

  Scenario: Two People Edit the Same Task at the Same Time
    When the first person tries to change the task's title to "A" many times
    And the second person tries to change the task's title to "B" many times
    Then both people should receive confirmation that their changes were saved
    And the final title of the task will be a random mix of the two versions