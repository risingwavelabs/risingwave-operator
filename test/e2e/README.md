# Bash Scripts for E2E Tests

## Code Style

Follow the Google [Shell Style Guide](https://google.github.io/styleguide/shellguide.html).

## File Structure

- `common`, directory that contains functions for common usages, e.g., `kubectl` and `awscli` operations.
- `testenv`, directory that contains functions for setting up and tearing down the environment to run the tests.
- `tests`, directory that contains test cases.
- `env`, file contains configuration envs.
- `e2e.sh`, entrance of the end-to-end test.

## Convention

- All test case functions should be start with `test::run` 