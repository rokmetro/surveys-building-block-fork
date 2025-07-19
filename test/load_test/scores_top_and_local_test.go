package load_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

// TestConfig holds configuration for load testing
type TestConfig struct {
	BaseURL         string
	AuthToken       string
	ConcurrentUsers int
	RequestsPerUser int
	TestDurationSec int
	RampUpTimeSec   int
}

// TestResult holds metrics for a single request
type TestResult struct {
	StatusCode   int
	ResponseTime time.Duration
	Success      bool
	ErrorMessage string
	ResponseSize int
}

// LoadTestMetrics aggregates all test results
type LoadTestMetrics struct {
	TotalRequests     int
	SuccessfulReqs    int
	FailedReqs        int
	AvgResponseTime   time.Duration
	MinResponseTime   time.Duration
	MaxResponseTime   time.Duration
	P95ResponseTime   time.Duration
	P99ResponseTime   time.Duration
	RequestsPerSecond float64
	ErrorRate         float64
	TotalDataReceived int64
}

// ScoreResponse represents the expected API response structure
type ScoreResponse struct {
	ID                     string    `json:"id"`
	OrgID                  string    `json:"org_id"`
	AppID                  string    `json:"app_id"`
	UserID                 string    `json:"user_id"`
	SurveyType             string    `json:"survey_type"`
	ExternalProfileID      string    `json:"external_profile_id"`
	Score                  float64   `json:"score"`
	ResponseCount          int       `json:"response_count"`
	PrevSurveyResponseDate time.Time `json:"prev_survey_response_date"`
	CurrentStreak          int       `json:"current_streak"`
	StreakMultiplier       float64   `json:"streak_multiplier"`
	AnswerCount            int       `json:"answer_count"`
	CorrectAnswerCount     int       `json:"correct_answer_count"`
	Rank                   int       `json:"rank"`
}

// TestScenario defines different test scenarios with varying parameters
type TestScenario struct {
	Name        string
	QueryParams map[string]string
}

// Default test configuration
var defaultConfig = TestConfig{
	BaseURL:         "http://localhost/surveys/api",
	AuthToken:       "", // Will be loaded from environment variable
	ConcurrentUsers: 50,
	RequestsPerUser: 20,
	TestDurationSec: 60,
	RampUpTimeSec:   10,
}

// Test scenarios with different parameter combinations
var testScenarios = []TestScenario{
	{
		Name: "Default Parameters",
		QueryParams: map[string]string{
			"limit":             "10",
			"offset":            "0",
			"local_limit":       "4",
			"above_pivot_limit": "2",
			"below_pivot_limit": "2",
		},
	},
	{
		Name: "High Limit",
		QueryParams: map[string]string{
			"limit":             "100",
			"offset":            "0",
			"local_limit":       "30",
			"above_pivot_limit": "25",
			"below_pivot_limit": "25",
		},
	},
	{
		Name: "Pagination Test",
		QueryParams: map[string]string{
			"limit":             "10",
			"offset":            "50",
			"local_limit":       "4",
			"above_pivot_limit": "2",
			"below_pivot_limit": "2",
		},
	},
	{
		Name: "Local Scores Focus",
		QueryParams: map[string]string{
			"limit":             "20",
			"local_limit":       "50",
			"above_pivot_limit": "10",
			"below_pivot_limit": "10",
		},
	},
}

// makeRequest performs a single HTTP request to the API
func makeRequest(config TestConfig, scenario TestScenario) TestResult {
	start := time.Now()

	// Build URL with query parameters
	url := fmt.Sprintf("%s/scores/top-and-local", config.BaseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return TestResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to create request: %v", err),
			ResponseTime: time.Since(start),
		}
	}

	// Add query parameters
	q := req.URL.Query()
	for key, value := range scenario.QueryParams {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	// Add authorization header
	if config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+config.AuthToken)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return TestResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Request failed: %v", err),
			ResponseTime: time.Since(start),
		}
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	responseTime := time.Since(start)

	if err != nil {
		return TestResult{
			StatusCode:   resp.StatusCode,
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to read response: %v", err),
			ResponseTime: responseTime,
		}
	}

	// Validate response structure for successful requests
	success := resp.StatusCode == 200
	errorMsg := ""

	if success {
		var scores []ScoreResponse
		if err := json.Unmarshal(body, &scores); err != nil {
			success = false
			errorMsg = fmt.Sprintf("Invalid JSON response: %v", err)
		}
	} else {
		errorMsg = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return TestResult{
		StatusCode:   resp.StatusCode,
		Success:      success,
		ErrorMessage: errorMsg,
		ResponseTime: responseTime,
		ResponseSize: len(body),
	}
}

// runLoadTest executes the load test for a specific scenario
func runLoadTest(t *testing.T, config TestConfig, scenario TestScenario) LoadTestMetrics {
	t.Logf("Running load test: %s", scenario.Name)
	t.Logf("Config: %d concurrent users, %d requests per user", config.ConcurrentUsers, config.RequestsPerUser)

	var wg sync.WaitGroup
	resultsChan := make(chan TestResult, config.ConcurrentUsers*config.RequestsPerUser)

	startTime := time.Now()

	// Launch concurrent users with ramp-up
	rampUpDelay := time.Duration(config.RampUpTimeSec) * time.Second / time.Duration(config.ConcurrentUsers)

	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)

		go func(userID int) {
			defer wg.Done()

			// Ramp-up delay
			time.Sleep(time.Duration(userID) * rampUpDelay)

			// Execute requests for this user
			for j := 0; j < config.RequestsPerUser; j++ {
				result := makeRequest(config, scenario)
				resultsChan <- result

				// Small delay between requests from same user
				time.Sleep(100 * time.Millisecond)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(resultsChan)

	// Collect and analyze results
	return analyzeResults(t, resultsChan, startTime)
}

// analyzeResults processes test results and calculates metrics
func analyzeResults(t *testing.T, resultsChan chan TestResult, startTime time.Time) LoadTestMetrics {
	var results []TestResult
	var totalDataReceived int64

	// Collect all results
	for result := range resultsChan {
		results = append(results, result)
		totalDataReceived += int64(result.ResponseSize)
	}

	if len(results) == 0 {
		return LoadTestMetrics{}
	}

	// Calculate basic metrics
	totalRequests := len(results)
	successfulReqs := 0
	var responseTimes []time.Duration

	for _, result := range results {
		if result.Success {
			successfulReqs++
		}
		responseTimes = append(responseTimes, result.ResponseTime)
	}

	failedReqs := totalRequests - successfulReqs
	errorRate := float64(failedReqs) / float64(totalRequests) * 100

	// Calculate response time statistics
	avgResponseTime := calculateAverage(responseTimes)
	minResponseTime := calculateMin(responseTimes)
	maxResponseTime := calculateMax(responseTimes)
	p95ResponseTime := calculatePercentile(responseTimes, 95)
	p99ResponseTime := calculatePercentile(responseTimes, 99)

	// Calculate requests per second
	totalDuration := time.Since(startTime)
	requestsPerSecond := float64(totalRequests) / totalDuration.Seconds()

	metrics := LoadTestMetrics{
		TotalRequests:     totalRequests,
		SuccessfulReqs:    successfulReqs,
		FailedReqs:        failedReqs,
		AvgResponseTime:   avgResponseTime,
		MinResponseTime:   minResponseTime,
		MaxResponseTime:   maxResponseTime,
		P95ResponseTime:   p95ResponseTime,
		P99ResponseTime:   p99ResponseTime,
		RequestsPerSecond: requestsPerSecond,
		ErrorRate:         errorRate,
		TotalDataReceived: totalDataReceived,
	}

	// Log results
	t.Logf("=== Load Test Results ===")
	t.Logf("Total Requests: %d", metrics.TotalRequests)
	t.Logf("Successful: %d (%.2f%%)", metrics.SuccessfulReqs, float64(metrics.SuccessfulReqs)/float64(metrics.TotalRequests)*100)
	t.Logf("Failed: %d (%.2f%%)", metrics.FailedReqs, metrics.ErrorRate)
	t.Logf("Requests/sec: %.2f", metrics.RequestsPerSecond)
	t.Logf("Avg Response Time: %v", metrics.AvgResponseTime)
	t.Logf("Min Response Time: %v", metrics.MinResponseTime)
	t.Logf("Max Response Time: %v", metrics.MaxResponseTime)
	t.Logf("95th Percentile: %v", metrics.P95ResponseTime)
	t.Logf("99th Percentile: %v", metrics.P99ResponseTime)
	t.Logf("Total Data Received: %.2f KB", float64(metrics.TotalDataReceived)/1024)

	// Log failed requests details
	if failedReqs > 0 {
		t.Logf("=== Error Details ===")
		errorCounts := make(map[string]int)
		for _, result := range results {
			if !result.Success {
				errorCounts[result.ErrorMessage]++
			}
		}
		for error, count := range errorCounts {
			t.Logf("Error: %s (Count: %d)", error, count)
		}
	}

	return metrics
}

// Helper functions for statistical calculations
func calculateAverage(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total / time.Duration(len(durations))
}

func calculateMin(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	min := durations[0]
	for _, d := range durations {
		if d < min {
			min = d
		}
	}
	return min
}

func calculateMax(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	max := durations[0]
	for _, d := range durations {
		if d > max {
			max = d
		}
	}
	return max
}

func calculatePercentile(durations []time.Duration, percentile int) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	// Simple bubble sort for small datasets
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)

	for i := 0; i < len(sorted); i++ {
		for j := 0; j < len(sorted)-1-i; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	index := int(float64(len(sorted)) * float64(percentile) / 100.0)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

// Test functions

// TestTopAndLocalScoresLoadBasic runs a basic load test with default parameters
func TestTopAndLocalScoresLoadBasic(t *testing.T) {
	config := getTestConfigFromEnv()

	// Check if auth token is available
	if config.AuthToken == "" {
		t.Skip("Skipping test: LOAD_TEST_AUTH_TOKEN environment variable not set")
	}

	config.ConcurrentUsers = 10
	config.RequestsPerUser = 5

	scenario := testScenarios[0] // Default parameters
	metrics := runLoadTest(t, config, scenario)

	// Assert basic success criteria
	if metrics.ErrorRate > 5.0 {
		t.Errorf("Error rate too high: %.2f%% (expected < 5%%)", metrics.ErrorRate)
	}

	if metrics.AvgResponseTime > 2*time.Second {
		t.Errorf("Average response time too high: %v (expected < 2s)", metrics.AvgResponseTime)
	}

	if metrics.RequestsPerSecond < 1.0 {
		t.Errorf("Throughput too low: %.2f req/s (expected > 1 req/s)", metrics.RequestsPerSecond)
	}
}

// TestTopAndLocalScoresLoadMedium runs a medium intensity load test
func TestTopAndLocalScoresLoadMedium(t *testing.T) {
	config := getTestConfigFromEnv()

	// Check if auth token is available
	if config.AuthToken == "" {
		t.Skip("Skipping test: LOAD_TEST_AUTH_TOKEN environment variable not set")
	}

	config.ConcurrentUsers = 25
	config.RequestsPerUser = 10

	scenario := testScenarios[1] // High limit scenario
	metrics := runLoadTest(t, config, scenario)

	// Assert performance criteria
	if metrics.ErrorRate > 10.0 {
		t.Errorf("Error rate too high: %.2f%% (expected < 10%%)", metrics.ErrorRate)
	}

	if metrics.P95ResponseTime > 5*time.Second {
		t.Errorf("95th percentile response time too high: %v (expected < 5s)", metrics.P95ResponseTime)
	}
}

// TestTopAndLocalScoresLoadHigh runs a high intensity load test
func TestTopAndLocalScoresLoadHigh(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high load test in short mode")
	}

	config := getTestConfigFromEnv()

	// Check if auth token is available
	if config.AuthToken == "" {
		t.Skip("Skipping test: LOAD_TEST_AUTH_TOKEN environment variable not set")
	}

	config.ConcurrentUsers = 50
	config.RequestsPerUser = 20
	config.RampUpTimeSec = 15

	scenario := testScenarios[0] // Default parameters
	metrics := runLoadTest(t, config, scenario)

	// Assert performance criteria for high load
	if metrics.ErrorRate > 15.0 {
		t.Errorf("Error rate too high: %.2f%% (expected < 15%%)", metrics.ErrorRate)
	}

	if metrics.P99ResponseTime > 10*time.Second {
		t.Errorf("99th percentile response time too high: %v (expected < 10s)", metrics.P99ResponseTime)
	}
}

// TestTopAndLocalScoresAllScenarios tests all parameter combinations
func TestTopAndLocalScoresAllScenarios(t *testing.T) {
	config := getTestConfigFromEnv()

	// Check if auth token is available
	if config.AuthToken == "" {
		t.Skip("Skipping test: LOAD_TEST_AUTH_TOKEN environment variable not set")
	}

	config.ConcurrentUsers = 15
	config.RequestsPerUser = 8

	for _, scenario := range testScenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			metrics := runLoadTest(t, config, scenario)

			// Basic assertions for all scenarios
			if metrics.TotalRequests == 0 {
				t.Error("No requests were made")
			}

			if metrics.ErrorRate > 20.0 {
				t.Errorf("Error rate too high for scenario %s: %.2f%%", scenario.Name, metrics.ErrorRate)
			}
		})
	}
}

// TestTopAndLocalScoresPagination specifically tests pagination performance
func TestTopAndLocalScoresPagination(t *testing.T) {
	config := getTestConfigFromEnv()

	// Check if auth token is available
	if config.AuthToken == "" {
		t.Skip("Skipping test: LOAD_TEST_AUTH_TOKEN environment variable not set")
	}

	config.ConcurrentUsers = 20
	config.RequestsPerUser = 5

	// Test different pagination scenarios
	paginationScenarios := []TestScenario{
		{
			Name: "First Page",
			QueryParams: map[string]string{
				"limit":  "20",
				"offset": "0",
			},
		},
		{
			Name: "Middle Page",
			QueryParams: map[string]string{
				"limit":  "20",
				"offset": "100",
			},
		},
		{
			Name: "High Offset",
			QueryParams: map[string]string{
				"limit":  "20",
				"offset": "500",
			},
		},
	}

	for _, scenario := range paginationScenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			metrics := runLoadTest(t, config, scenario)

			// Pagination-specific assertions
			if metrics.ErrorRate > 10.0 {
				t.Errorf("Error rate too high for %s: %.2f%%", scenario.Name, metrics.ErrorRate)
			}

			// Higher offsets might be slower, but should still be reasonable
			maxExpectedTime := 3 * time.Second
			if scenario.Name == "High Offset" {
				maxExpectedTime = 5 * time.Second
			}

			if metrics.AvgResponseTime > maxExpectedTime {
				t.Errorf("Average response time too high for %s: %v (expected < %v)",
					scenario.Name, metrics.AvgResponseTime, maxExpectedTime)
			}
		})
	}
}

// BenchmarkTopAndLocalScoresAPI provides benchmark results
func BenchmarkTopAndLocalScoresAPI(b *testing.B) {
	config := getTestConfigFromEnv()

	// Check if auth token is available
	if config.AuthToken == "" {
		b.Skip("Skipping benchmark: LOAD_TEST_AUTH_TOKEN environment variable not set")
	}

	scenario := testScenarios[0] // Default parameters

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result := makeRequest(config, scenario)
			if !result.Success {
				b.Errorf("Request failed: %s", result.ErrorMessage)
			}
		}
	})
}

// Helper function to create a test configuration from environment variables
func getTestConfigFromEnv() TestConfig {
	config := defaultConfig

	// Load auth token from environment variable (required)
	if authToken := os.Getenv("LOAD_TEST_AUTH_TOKEN"); authToken != "" {
		config.AuthToken = authToken
	}

	// Override other settings with environment variables if available
	if baseURL := os.Getenv("LOAD_TEST_BASE_URL"); baseURL != "" {
		config.BaseURL = baseURL
	}

	if concurrentUsers := os.Getenv("LOAD_TEST_CONCURRENT_USERS"); concurrentUsers != "" {
		if users, err := strconv.Atoi(concurrentUsers); err == nil && users > 0 {
			config.ConcurrentUsers = users
		}
	}

	if requestsPerUser := os.Getenv("LOAD_TEST_REQUESTS_PER_USER"); requestsPerUser != "" {
		if requests, err := strconv.Atoi(requestsPerUser); err == nil && requests > 0 {
			config.RequestsPerUser = requests
		}
	}

	if testDuration := os.Getenv("LOAD_TEST_DURATION_SEC"); testDuration != "" {
		if duration, err := strconv.Atoi(testDuration); err == nil && duration > 0 {
			config.TestDurationSec = duration
		}
	}

	if rampUpTime := os.Getenv("LOAD_TEST_RAMP_UP_SEC"); rampUpTime != "" {
		if rampUp, err := strconv.Atoi(rampUpTime); err == nil && rampUp >= 0 {
			config.RampUpTimeSec = rampUp
		}
	}

	return config
}
