package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setEnv(t *testing.T, key, value string) {
	t.Setenv(key, value)
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		expected     string
	}{
		{
			name:         "returns env var value when set",
			key:          "TEST_GET_ENV_SET",
			defaultValue: "fallback",
			envValue:     "from_env",
			setEnv:       true,
			expected:     "from_env",
		},
		{
			name:         "returns default when env var is empty",
			key:          "TEST_GET_ENV_EMPTY",
			defaultValue: "fallback",
			envValue:     "",
			setEnv:       false,
			expected:     "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				setEnv(t, tt.key, tt.envValue)
			}

			result := getEnv(tt.key, tt.defaultValue)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		setEnv       bool
		expected     int
	}{
		{
			name:         "returns parsed int when valid",
			key:          "TEST_INT_VALID",
			defaultValue: 10,
			envValue:     "42",
			setEnv:       true,
			expected:     42,
		},
		{
			name:         "returns default when env var is empty",
			key:          "TEST_INT_EMPTY",
			defaultValue: 10,
			setEnv:       false,
			expected:     10,
		},
		{
			name:         "returns default when env var is not a valid int",
			key:          "TEST_INT_INVALID",
			defaultValue: 10,
			envValue:     "not_a_number",
			setEnv:       true,
			expected:     10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				setEnv(t, tt.key, tt.envValue)
			}

			result := getEnvAsInt(tt.key, tt.defaultValue)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvAsDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue time.Duration
		envValue     string
		setEnv       bool
		expected     time.Duration
	}{
		{
			name:         "returns parsed duration when valid",
			key:          "TEST_DUR_VALID",
			defaultValue: 5 * time.Second,
			envValue:     "30s",
			setEnv:       true,
			expected:     30 * time.Second,
		},
		{
			name:         "returns default when env var is empty",
			key:          "TEST_DUR_EMPTY",
			defaultValue: 5 * time.Second,
			setEnv:       false,
			expected:     5 * time.Second,
		},
		{
			name:         "returns default when env var is not a valid duration",
			key:          "TEST_DUR_INVALID",
			defaultValue: 5 * time.Second,
			envValue:     "not_a_duration",
			setEnv:       true,
			expected:     5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				setEnv(t, tt.key, tt.envValue)
			}

			result := getEnvAsDuration(tt.key, tt.defaultValue)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvAsSlice(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue []string
		envValue     string
		setEnv       bool
		expected     []string
	}{
		{
			name:         "returns single element slice for single value",
			key:          "TEST_SLICE_SINGLE",
			defaultValue: nil,
			envValue:     "http://es:9200",
			setEnv:       true,
			expected:     []string{"http://es:9200"},
		},
		{
			name:         "returns multiple elements for comma-separated values",
			key:          "TEST_SLICE_MULTI",
			defaultValue: nil,
			envValue:     "http://es1:9200,http://es2:9200,http://es3:9200",
			setEnv:       true,
			expected:     []string{"http://es1:9200", "http://es2:9200", "http://es3:9200"},
		},
		{
			name:         "trims whitespace from elements",
			key:          "TEST_SLICE_SPACES",
			defaultValue: nil,
			envValue:     " http://es1:9200 , http://es2:9200 ",
			setEnv:       true,
			expected:     []string{"http://es1:9200", "http://es2:9200"},
		},
		{
			name:         "returns default when env var is empty",
			key:          "TEST_SLICE_EMPTY",
			defaultValue: []string{"http://default:9200"},
			setEnv:       false,
			expected:     []string{"http://default:9200"},
		},
		{
			name:         "returns default when env var contains only commas and spaces",
			key:          "TEST_SLICE_ONLY_COMMAS",
			defaultValue: []string{"http://default:9200"},
			envValue:     " , , , ",
			setEnv:       true,
			expected:     []string{"http://default:9200"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				setEnv(t, tt.key, tt.envValue)
			}

			result := getEnvAsSlice(tt.key, tt.defaultValue)

			assert.Equal(t, tt.expected, result)
		})
	}
}
