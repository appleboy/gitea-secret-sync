package main

import (
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with all fields",
			config: config{
				Server:     "https://gitea.example.com",
				Token:      "1234567890abcdef1234567890abcdef12345678",
				SkipVerify: false,
				Timeout:    60 * time.Second,
				RetryCount: 5,
			},
			wantErr: false,
		},
		{
			name: "valid config with minimal required fields",
			config: config{
				Server: "https://gitea.example.com",
				Token:  "1234567890abcdef1234567890abcdef12345678",
			},
			wantErr: false,
		},
		{
			name: "empty server URL",
			config: config{
				Server: "",
				Token:  "1234567890abcdef1234567890abcdef12345678",
			},
			wantErr: true,
			errMsg:  "server URL is required",
		},
		{
			name: "empty token",
			config: config{
				Server: "https://gitea.example.com",
				Token:  "",
			},
			wantErr: true,
			errMsg:  "token is required",
		},
		{
			name: "token too short",
			config: config{
				Server: "https://gitea.example.com",
				Token:  "short",
			},
			wantErr: true,
			errMsg:  "token appears invalid (expected 40 characters)",
		},
		{
			name: "token exactly 40 characters",
			config: config{
				Server: "https://gitea.example.com",
				Token:  "1234567890abcdef1234567890abcdef12345678",
			},
			wantErr: false,
		},
		{
			name: "token longer than 40 characters",
			config: config{
				Server: "https://gitea.example.com",
				Token:  "1234567890abcdef1234567890abcdef123456789extra",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("config.Validate() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("config.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("config.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestConfig_Validate_DefaultValues(t *testing.T) {
	tests := []struct {
		name               string
		config             config
		expectedTimeout    time.Duration
		expectedRetryCount int
	}{
		{
			name: "sets default timeout when zero",
			config: config{
				Server:     "https://gitea.example.com",
				Token:      "1234567890abcdef1234567890abcdef12345678",
				Timeout:    0,
				RetryCount: 5,
			},
			expectedTimeout:    30 * time.Second,
			expectedRetryCount: 5,
		},
		{
			name: "sets default retry count when zero",
			config: config{
				Server:     "https://gitea.example.com",
				Token:      "1234567890abcdef1234567890abcdef12345678",
				Timeout:    60 * time.Second,
				RetryCount: 0,
			},
			expectedTimeout:    60 * time.Second,
			expectedRetryCount: 3,
		},
		{
			name: "sets both defaults when both are zero",
			config: config{
				Server:     "https://gitea.example.com",
				Token:      "1234567890abcdef1234567890abcdef12345678",
				Timeout:    0,
				RetryCount: 0,
			},
			expectedTimeout:    30 * time.Second,
			expectedRetryCount: 3,
		},
		{
			name: "preserves non-zero values",
			config: config{
				Server:     "https://gitea.example.com",
				Token:      "1234567890abcdef1234567890abcdef12345678",
				Timeout:    120 * time.Second,
				RetryCount: 10,
			},
			expectedTimeout:    120 * time.Second,
			expectedRetryCount: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err != nil {
				t.Fatalf("config.Validate() unexpected error = %v", err)
			}

			if tt.config.Timeout != tt.expectedTimeout {
				t.Errorf("config.Validate() Timeout = %v, want %v", tt.config.Timeout, tt.expectedTimeout)
			}

			if tt.config.RetryCount != tt.expectedRetryCount {
				t.Errorf("config.Validate() RetryCount = %v, want %v", tt.config.RetryCount, tt.expectedRetryCount)
			}
		})
	}
}

func TestConfig_Validate_BoundaryValues(t *testing.T) {
	tests := []struct {
		name    string
		config  config
		wantErr bool
	}{
		{
			name: "token with exactly 39 characters (invalid)",
			config: config{
				Server: "https://gitea.example.com",
				Token:  "123456789012345678901234567890123456789", // 39 chars
			},
			wantErr: true,
		},
		{
			name: "token with exactly 40 characters (valid)",
			config: config{
				Server: "https://gitea.example.com",
				Token:  "1234567890123456789012345678901234567890", // 40 chars
			},
			wantErr: false,
		},
		{
			name: "token with 41 characters (valid)",
			config: config{
				Server: "https://gitea.example.com",
				Token:  "12345678901234567890123456789012345678901", // 41 chars
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr && err == nil {
				t.Errorf("config.Validate() expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("config.Validate() unexpected error = %v", err)
			}
		})
	}
}

// BenchmarkConfig_Validate benchmarks the Validate method
func BenchmarkConfig_Validate(b *testing.B) {
	config := config{
		Server: "https://gitea.example.com",
		Token:  "1234567890abcdef1234567890abcdef12345678",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}
