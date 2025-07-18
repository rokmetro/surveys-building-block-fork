
Usage Instructions:

1. Set up your test environment:
   - Ensure the surveys API is running
   - Obtain a valid JWT authentication token
   - Set the LOAD_TEST_AUTH_TOKEN environment variable:
     export LOAD_TEST_AUTH_TOKEN="your-jwt-token-here"

2. Optional environment variables for customization:
   - export LOAD_TEST_BASE_URL="http://localhost/surveys/api"
   - export LOAD_TEST_CONCURRENT_USERS="50"
   - export LOAD_TEST_REQUESTS_PER_USER="20"
   - export LOAD_TEST_DURATION_SEC="60"
   - export LOAD_TEST_RAMP_UP_SEC="10"

3. Run different test levels:

   - Basic load test (quick):
   go test -v -run TestTopAndLocalScoresLoadBasic

   - Medium load test:
   go test -v -run TestTopAndLocalScoresLoadMedium

   - High load test (takes longer):
   go test -v -run TestTopAndLocalScoresLoadHigh

   - All scenarios:
   go test -v -run TestTopAndLocalScoresAllScenarios

   - Pagination-specific tests:
   go test -v -run TestTopAndLocalScoresPagination

   - Benchmark:
   go test -v -bench=BenchmarkTopAndLocalScoresAPI

   - Run all tests:
   go test -v ./load_test/

4. Customize test parameters:
   - Use environment variables to adjust settings (see step 2)
   - Add new test scenarios to the testScenarios slice
   - Adjust assertion thresholds in test functions based on your performance requirements

5. Interpreting results:
   - Success rate should be > 95% for most scenarios
   - Average response time should be < 2 seconds for normal load
   - 95th percentile should be < 5 seconds
   - 99th percentile should be < 10 seconds
   - Requests per second indicates throughput capacity

6. Performance tuning:
   - If error rates are high, check server logs and database performance
   - If response times are high, consider database indexing and query optimization
   - Monitor server resources (CPU, memory, database connections) during tests