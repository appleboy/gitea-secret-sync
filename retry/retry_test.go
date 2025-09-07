package retry

import (
	"errors"
	"log/slog"
	"net/http"
	"testing"

	gsdk "code.gitea.io/sdk/gitea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock types for testing
type MockNetError struct {
	timeout bool
	temp    bool
}

func (e MockNetError) Error() string   { return "mock network error" }
func (e MockNetError) Timeout() bool   { return e.timeout }
func (e MockNetError) Temporary() bool { return e.temp }

// TestNewDefaultRetrier tests the generic retrier constructor
func TestNewDefaultRetrier(t *testing.T) {
	tests := []struct {
		name     string
		config   *RetryConfig
		expected *RetryConfig
	}{
		{
			name:   "nil config uses defaults",
			config: nil,
			expected: &RetryConfig{
				MaxRetries: 3,
				Logger:     slog.Default(),
			},
		},
		{
			name: "nil logger uses default",
			config: &RetryConfig{
				MaxRetries: 5,
				Logger:     nil,
			},
			expected: &RetryConfig{
				MaxRetries: 5,
				Logger:     slog.Default(),
			},
		},
		{
			name: "custom config preserved",
			config: &RetryConfig{
				MaxRetries: 2,
				Logger:     slog.New(slog.NewTextHandler(nil, nil)),
			},
			expected: &RetryConfig{
				MaxRetries: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrier := NewDefaultRetrier[string](tt.config)

			require.NotNil(t, retrier)
			require.NotNil(t, retrier.config)
			assert.Equal(t, tt.expected.MaxRetries, retrier.config.MaxRetries)

			if tt.expected.Logger != nil {
				assert.Equal(t, tt.expected.Logger, retrier.config.Logger)
			} else {
				assert.NotNil(t, retrier.config.Logger)
			}
		})
	}
}

// TestNewGiteaRetrier tests the Gitea-specific retrier constructor
func TestNewGiteaRetrier(t *testing.T) {
	config := &RetryConfig{
		MaxRetries: 5,
		Logger:     slog.Default(),
	}

	retrier := NewGiteaRetrier(config)

	require.NotNil(t, retrier)
	require.NotNil(t, retrier.config)
	assert.Equal(t, 5, retrier.config.MaxRetries)
}

// TestDefaultRetrier_DoWithRetry tests the generic retry logic
func TestDefaultRetrier_DoWithRetry(t *testing.T) {
	// Note: We can't mock time.Sleep directly, but tests should be fast anyway
	// since successful operations don't trigger sleep

	tests := []struct {
		name           string
		maxRetries     int
		operations     []func() (string, error)
		expectedResult string
		expectedError  string
		expectedCalls  int
	}{
		{
			name:       "success on first try",
			maxRetries: 3,
			operations: []func() (string, error){
				func() (string, error) { return "success", nil },
			},
			expectedResult: "success",
			expectedCalls:  1,
		},
		{
			name:       "success after retry",
			maxRetries: 3,
			operations: []func() (string, error){
				func() (string, error) { return "", errors.New("500 server error") },
				func() (string, error) { return "success", nil },
			},
			expectedResult: "success",
			expectedCalls:  2,
		},
		{
			name:       "failure after all retries",
			maxRetries: 2,
			operations: []func() (string, error){
				func() (string, error) { return "", errors.New("500 server error") },
				func() (string, error) { return "", errors.New("500 server error") },
			},
			expectedResult: "",
			expectedError:  "operation failed after 2 retries",
			expectedCalls:  2,
		},
		{
			name:       "non-retryable error stops immediately",
			maxRetries: 3,
			operations: []func() (string, error){
				func() (string, error) { return "", errors.New("400 bad request") },
			},
			expectedResult: "",
			expectedError:  "400 bad request",
			expectedCalls:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &RetryConfig{
				MaxRetries: tt.maxRetries,
				Logger:     slog.Default(),
			}

			retrier := NewDefaultRetrier[string](config)
			callCount := 0

			operation := func() (string, error) {
				if callCount < len(tt.operations) {
					result := tt.operations[callCount]
					callCount++
					return result()
				}
				return tt.operations[len(tt.operations)-1]()
			}

			result, err := retrier.DoWithRetry(operation)

			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedCalls, callCount)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDefaultRetrier_IsRetryable tests the default retry logic
func TestDefaultRetrier_IsRetryable(t *testing.T) {
	retrier := NewDefaultRetrier[string](&RetryConfig{
		MaxRetries: 3,
		Logger:     slog.Default(),
	})

	tests := []struct {
		name     string
		resp     string
		err      error
		expected bool
	}{
		{
			name:     "no error means success, no retry",
			resp:     "success",
			err:      nil,
			expected: false,
		},
		{
			name:     "network timeout error is retryable",
			resp:     "",
			err:      MockNetError{timeout: true},
			expected: true,
		},
		{
			name:     "server error 500 is retryable",
			resp:     "",
			err:      errors.New("HTTP 500 server error"),
			expected: true,
		},
		{
			name:     "server error 502 is retryable",
			resp:     "",
			err:      errors.New("502 bad gateway"),
			expected: true,
		},
		{
			name:     "server error 503 is retryable",
			resp:     "",
			err:      errors.New("503 service unavailable"),
			expected: true,
		},
		{
			name:     "server error 504 is retryable",
			resp:     "",
			err:      errors.New("504 gateway timeout"),
			expected: true,
		},
		{
			name: "connection refused is retryable",
			resp: "",
			err: errors.
				New("connection refused"),
			expected: true,
		},
		{
			name:     "network unreachable is retryable",
			resp:     "",
			err:      errors.New("network is unreachable"),
			expected: true,
		},
		{
			name:     "no such host is retryable",
			resp:     "",
			err:      errors.New("no such host"),
			expected: true,
		},
		{
			name:     "generic error is not retryable",
			resp:     "",
			err:      errors.New("generic error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := retrier.IsRetryable(tt.resp, tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGiteaRetrier_IsRetryable tests the Gitea-specific retry logic
func TestGiteaRetrier_IsRetryable(t *testing.T) {
	retrier := NewGiteaRetrier(&RetryConfig{
		MaxRetries: 3,
		Logger:     slog.Default(),
	})

	tests := []struct {
		name     string
		resp     *gsdk.Response
		err      error
		expected bool
	}{
		{
			name: "2xx success should not retry",
			resp: &gsdk.Response{
				Response: &http.Response{StatusCode: 200},
			},
			err:      nil,
			expected: false,
		},
		{
			name: "201 success should not retry",
			resp: &gsdk.Response{
				Response: &http.Response{StatusCode: 201},
			},
			err:      nil,
			expected: false,
		},
		{
			name: "500 server error should retry",
			resp: &gsdk.Response{
				Response: &http.Response{StatusCode: 500},
			},
			err:      nil,
			expected: true,
		},
		{
			name: "502 bad gateway should retry",
			resp: &gsdk.Response{
				Response: &http.Response{StatusCode: 502},
			},
			err:      nil,
			expected: true,
		},
		{
			name: "503 service unavailable should retry",
			resp: &gsdk.Response{
				Response: &http.Response{StatusCode: 503},
			},
			err:      nil,
			expected: true,
		},
		{
			name: "504 gateway timeout should retry",
			resp: &gsdk.Response{
				Response: &http.Response{StatusCode: 504},
			},
			err:      nil,
			expected: true,
		},
		{
			name: "408 request timeout should retry",
			resp: &gsdk.Response{
				Response: &http.Response{StatusCode: 408},
			},
			err:      nil,
			expected: true,
		},
		{
			name: "429 too many requests should retry",
			resp: &gsdk.Response{
				Response: &http.Response{StatusCode: 429},
			},
			err:      nil,
			expected: true,
		},
		{
			name: "401 unauthorized should not retry",
			resp: &gsdk.Response{
				Response: &http.Response{StatusCode: 401},
			},
			err:      nil,
			expected: false,
		},
		{
			name: "403 forbidden should not retry",
			resp: &gsdk.Response{
				Response: &http.Response{StatusCode: 403},
			},
			err:      nil,
			expected: false,
		},
		{
			name: "404 not found should not retry",
			resp: &gsdk.Response{
				Response: &http.Response{StatusCode: 404},
			},
			err:      nil,
			expected: false,
		},
		{
			name: "422 unprocessable entity should not retry",
			resp: &gsdk.Response{
				Response: &http.Response{StatusCode: 422},
			},
			err:      nil,
			expected: false,
		},
		{
			name:     "network timeout with error should retry",
			resp:     nil,
			err:      MockNetError{timeout: true},
			expected: true,
		},
		{
			name:     "500 error in message should retry",
			resp:     nil,
			err:      errors.New("HTTP 500 internal server error"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := retrier.IsRetryable(tt.resp, tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGiteaRetrier_DoWithRetry tests the Gitea retrier with actual response types
func TestGiteaRetrier_DoWithRetry(t *testing.T) {
	// Note: We can't mock time.Sleep directly, but tests should be fast anyway
	// since successful operations don't trigger sleep

	config := &RetryConfig{
		MaxRetries: 2,
		Logger:     slog.Default(),
	}
	retrier := NewGiteaRetrier(config)

	t.Run("success after retry", func(t *testing.T) {
		callCount := 0
		operation := func() (*gsdk.Response, error) {
			callCount++
			if callCount == 1 {
				return &gsdk.Response{
					Response: &http.Response{StatusCode: 500},
				}, nil
			}
			return &gsdk.Response{
				Response: &http.Response{StatusCode: 200},
			}, nil
		}

		result, err := retrier.DoWithRetry(operation)

		require.NoError(t, err)
		assert.Equal(t, 200, result.StatusCode)
		assert.Equal(t, 2, callCount)
	})

	t.Run("non-retryable error stops immediately", func(t *testing.T) {
		callCount := 0
		operation := func() (*gsdk.Response, error) {
			callCount++
			return &gsdk.Response{
				Response: &http.Response{StatusCode: 401},
			}, nil
		}

		result, err := retrier.DoWithRetry(operation)

		require.NoError(t, err)
		assert.Equal(t, 401, result.StatusCode)
		assert.Equal(t, 1, callCount)
	})
}

// TestGenericRetrier_DifferentTypes tests the generic retrier with different types
func TestGenericRetrier_DifferentTypes(t *testing.T) {
	// Note: We can't mock time.Sleep directly, but tests should be fast anyway
	// since successful operations don't trigger sleep

	config := &RetryConfig{
		MaxRetries: 2,
		Logger:     slog.Default(),
	}

	t.Run("string retrier", func(t *testing.T) {
		retrier := NewDefaultRetrier[string](config)

		result, err := retrier.DoWithRetry(func() (string, error) {
			return "test", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("int retrier", func(t *testing.T) {
		retrier := NewDefaultRetrier[int](config)

		result, err := retrier.DoWithRetry(func() (int, error) {
			return 42, nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("struct retrier", func(t *testing.T) {
		type TestStruct struct {
			Name  string
			Value int
		}

		retrier := NewDefaultRetrier[*TestStruct](config)

		expected := &TestStruct{Name: "test", Value: 123}
		result, err := retrier.DoWithRetry(func() (*TestStruct, error) {
			return expected, nil
		})

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}

// Benchmark tests
func BenchmarkDefaultRetrier_DoWithRetry_Success(b *testing.B) {
	config := &RetryConfig{
		MaxRetries: 3,
		Logger:     slog.Default(),
	}
	retrier := NewDefaultRetrier[string](config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = retrier.DoWithRetry(func() (string, error) {
			return "success", nil
		})
	}
}

func BenchmarkGiteaRetrier_DoWithRetry_Success(b *testing.B) {
	config := &RetryConfig{
		MaxRetries: 3,
		Logger:     slog.Default(),
	}
	retrier := NewGiteaRetrier(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = retrier.DoWithRetry(func() (*gsdk.Response, error) {
			return &gsdk.Response{
				Response: &http.Response{StatusCode: 200},
			}, nil
		})
	}
}
