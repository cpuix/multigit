package errors_test

import (
	"fmt"
	"testing"

	"github.com/cpuix/multigit/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Test creating a new error
	err := errors.New(errors.ErrorTypeConfig, "test error")
	
	// Verify error properties
	assert.Equal(t, errors.ErrorTypeConfig, err.Type, "Error type should match")
	assert.Equal(t, "test error", err.Error(), "Error message should match")
	assert.NotNil(t, err.Context, "Context should be initialized")
	assert.Empty(t, err.Context, "Context should be empty")
}

func TestWrap(t *testing.T) {
	// Create an original error
	originalErr := fmt.Errorf("original error")
	
	// Wrap the error
	wrappedErr := errors.Wrap(originalErr, errors.ErrorTypeSSH, "wrapped message")
	
	// Verify error properties
	assert.Equal(t, errors.ErrorTypeSSH, wrappedErr.Type, "Error type should match")
	assert.Contains(t, wrappedErr.Error(), "original error", "Error should contain original message")
	assert.Contains(t, wrappedErr.Error(), "wrapped message", "Error should contain wrapped message")
	assert.NotNil(t, wrappedErr.Context, "Context should be initialized")
	assert.Empty(t, wrappedErr.Context, "Context should be empty")
	
	// Test unwrapping
	unwrappedErr := wrappedErr.Unwrap()
	assert.NotNil(t, unwrappedErr, "Unwrapped error should not be nil")
	assert.Contains(t, unwrappedErr.Error(), "original error", "Unwrapped error should contain original message")
}

func TestErrorf(t *testing.T) {
	// Test creating a formatted error
	err := errors.Errorf(errors.ErrorTypeGit, "formatted %s: %d", "error", 42)
	
	// Verify error properties
	assert.Equal(t, errors.ErrorTypeGit, err.Type, "Error type should match")
	assert.Equal(t, "formatted error: 42", err.Error(), "Error message should match formatted string")
	assert.NotNil(t, err.Context, "Context should be initialized")
	assert.Empty(t, err.Context, "Context should be empty")
}

func TestWithContext(t *testing.T) {
	// Create an error with context
	err := errors.New(errors.ErrorTypeIO, "test error")
	err = err.WithContext("key1", "value1").
		WithContext("key2", 42)
	
	// Verify context
	assert.Equal(t, 2, len(err.Context), "Context should have 2 entries")
	assert.Equal(t, "value1", err.Context["key1"], "Context key1 should match")
	assert.Equal(t, 42, err.Context["key2"], "Context key2 should match")
}

func TestIsType(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		errType  errors.ErrorType
		expected bool
	}{
		{
			name:     "Matching error type",
			err:      errors.New(errors.ErrorTypeConfig, "config error"),
			errType:  errors.ErrorTypeConfig,
			expected: true,
		},
		{
			name:     "Non-matching error type",
			err:      errors.New(errors.ErrorTypeConfig, "config error"),
			errType:  errors.ErrorTypeSSH,
			expected: false,
		},
		{
			name:     "Standard error",
			err:      fmt.Errorf("standard error"),
			errType:  errors.ErrorTypeConfig,
			expected: false,
		},
		{
			name:     "Wrapped error with matching type",
			err:      errors.Wrap(fmt.Errorf("original"), errors.ErrorTypeGit, "wrapped"),
			errType:  errors.ErrorTypeGit,
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.IsType(tt.err, tt.errType)
			assert.Equal(t, tt.expected, result, "IsType result should match expected value")
		})
	}
}

func TestGetContext(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectedEmpty bool
	}{
		{
			name:          "Error with context",
			err:           errors.New(errors.ErrorTypeConfig, "error").WithContext("key", "value"),
			expectedEmpty: false,
		},
		{
			name:          "Error without context",
			err:           errors.New(errors.ErrorTypeConfig, "error"),
			expectedEmpty: true,
		},
		{
			name:          "Standard error",
			err:           fmt.Errorf("standard error"),
			expectedEmpty: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := errors.GetContext(tt.err)
			if tt.expectedEmpty {
				assert.Empty(t, context, "Context should be empty or nil")
			} else {
				assert.NotEmpty(t, context, "Context should not be empty")
			}
		})
	}
}

func TestErrorTypes(t *testing.T) {
	// Test that all error types are defined
	assert.Equal(t, errors.ErrorType("config"), errors.ErrorTypeConfig, "ErrorTypeConfig should be 'config'")
	assert.Equal(t, errors.ErrorType("ssh"), errors.ErrorTypeSSH, "ErrorTypeSSH should be 'ssh'")
	assert.Equal(t, errors.ErrorType("git"), errors.ErrorTypeGit, "ErrorTypeGit should be 'git'")
	assert.Equal(t, errors.ErrorType("io"), errors.ErrorTypeIO, "ErrorTypeIO should be 'io'")
	assert.Equal(t, errors.ErrorType("validation"), errors.ErrorTypeValidation, "ErrorTypeValidation should be 'validation'")
}
