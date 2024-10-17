# Testing Instructions for Baxfer

## Running Tests

1. Run all tests:
   ```
   go test ./...
   ```

2. Run tests with verbose output:
   ```
   go test -v ./...
   ```

3. Run tests for a specific package:
   ```
   go test -v ./pkg/storage
   ```

4. Run a specific test:
   ```
   go test -v ./pkg/storage -run TestUpload
   ```

5. Run tests with coverage:
   ```
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out
   ```

## Writing New Tests

1. Create test files with the naming convention `*_test.go`.
2. Use table-driven tests for functions with multiple scenarios.
3. Use mocks for external dependencies (like uploaders and loggers).
4. Aim for high code coverage, especially for critical functions.

## Continuous Integration

1. Set up GitHub Actions to run tests on every push and pull request.
2. Configure the CI to report test coverage.

## Best Practices

1. Write tests before implementing new features (Test-Driven Development).
2. Update tests when modifying existing functionality.
3. Use meaningful test names that describe the scenario being tested.
4. Keep tests independent and idempotent.
5. Use test fixtures for complex input data.
6. Regularly run the entire test suite locally before pushing changes.

## Maintaining Tests

1. Regularly review and update tests to ensure they remain relevant.
2. Remove or update obsolete tests when functionality changes.
3. Refactor tests to reduce duplication and improve readability.
4. Periodically run tests with race detection:
   ```
   go test -race ./...
   ```

## Integration Tests

1. Create a separate package for integration tests (e.g., `integrationtest`).
2. Use environment variables to control whether integration tests run:
   ```
   BAXFER_INTEGRATION_TESTS=1 go test ./integrationtest
   ```
3. Ensure integration tests clean up any resources they create.

## Performance Tests

1. Use benchmarks for performance-critical code:
   ```go
   func BenchmarkUpload(b *testing.B) {
       // Setup
       for i := 0; i < b.N; i++ {
           // Run the function being benchmarked
       }
   }
   ```
2. Run benchmarks:
   ```
   go test -bench=. ./...
   ```

Remember to update this document as the project evolves and new testing requirements emerge.
