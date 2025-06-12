package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestResourceNotFoundError(t *testing.T) {
	tests := []struct {
		name        string
		err         *ResourceNotFoundError
		wantMessage string
		targetError error
		shouldMatch bool
	}{
		{
			name: "container not found",
			err: &ResourceNotFoundError{
				ResourceType: "container",
				ID:           "abc123",
			},
			wantMessage: "container not found: abc123",
			targetError: ErrNotFound,
			shouldMatch: true,
		},
		{
			name: "image not found",
			err: &ResourceNotFoundError{
				ResourceType: "image",
				ID:           "nginx:latest",
			},
			wantMessage: "image not found: nginx:latest",
			targetError: ErrNotFound,
			shouldMatch: true,
		},
		{
			name: "does not match other error",
			err: &ResourceNotFoundError{
				ResourceType: "container",
				ID:           "abc123",
			},
			targetError: ErrAlreadyExists,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantMessage != "" && tt.err.Error() != tt.wantMessage {
				t.Errorf("ResourceNotFoundError.Error() = %v, want %v", tt.err.Error(), tt.wantMessage)
			}
			if errors.Is(tt.err, tt.targetError) != tt.shouldMatch {
				t.Errorf("errors.Is() match = %v, want %v", !tt.shouldMatch, tt.shouldMatch)
			}
		})
	}
}

func TestResourceExistsError(t *testing.T) {
	tests := []struct {
		name        string
		err         *ResourceExistsError
		wantMessage string
		targetError error
		shouldMatch bool
	}{
		{
			name: "container already exists",
			err: &ResourceExistsError{
				ResourceType: "container",
				ID:           "abc123",
			},
			wantMessage: "container already exists: abc123",
			targetError: ErrAlreadyExists,
			shouldMatch: true,
		},
		{
			name: "does not match other error",
			err: &ResourceExistsError{
				ResourceType: "container",
				ID:           "abc123",
			},
			targetError: ErrNotFound,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantMessage != "" && tt.err.Error() != tt.wantMessage {
				t.Errorf("ResourceExistsError.Error() = %v, want %v", tt.err.Error(), tt.wantMessage)
			}
			if errors.Is(tt.err, tt.targetError) != tt.shouldMatch {
				t.Errorf("errors.Is() match = %v, want %v", !tt.shouldMatch, tt.shouldMatch)
			}
		})
	}
}

func TestConfigError(t *testing.T) {
	tests := []struct {
		name        string
		err         *ConfigError
		wantMessage string
		targetError error
		shouldMatch bool
	}{
		{
			name: "invalid port configuration",
			err: &ConfigError{
				Field:   "port",
				Message: "must be between 1 and 65535",
			},
			wantMessage: "invalid configuration for port: must be between 1 and 65535",
			targetError: ErrInvalidConfig,
			shouldMatch: true,
		},
		{
			name: "does not match other error",
			err: &ConfigError{
				Field:   "port",
				Message: "invalid",
			},
			targetError: ErrNotFound,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantMessage != "" && tt.err.Error() != tt.wantMessage {
				t.Errorf("ConfigError.Error() = %v, want %v", tt.err.Error(), tt.wantMessage)
			}
			if errors.Is(tt.err, tt.targetError) != tt.shouldMatch {
				t.Errorf("errors.Is() match = %v, want %v", !tt.shouldMatch, tt.shouldMatch)
			}
		})
	}
}

func TestOperationalErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantMessage string
	}{
		{
			name: "container error",
			err: &ContainerError{
				ID:      "abc123",
				Op:      "start",
				Message: "port already in use",
			},
			wantMessage: "container abc123: start failed: port already in use",
		},
		{
			name: "network error",
			err: &NetworkError{
				ID:      "bridge",
				Op:      "create",
				Message: "address already in use",
			},
			wantMessage: "network bridge: create failed: address already in use",
		},
		{
			name: "volume error",
			err: &VolumeError{
				Name:    "data",
				Op:      "mount",
				Message: "permission denied",
			},
			wantMessage: "volume data: mount failed: permission denied",
		},
		{
			name: "image error",
			err: &ImageError{
				Ref:     "nginx:latest",
				Op:      "pull",
				Message: "registry unavailable",
			},
			wantMessage: "image nginx:latest: pull failed: registry unavailable",
		},
		{
			name: "exec error",
			err: &ExecError{
				ID:      "abc123",
				Op:      "start",
				Message: "command not found",
			},
			wantMessage: "exec abc123: start failed: command not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.wantMessage {
				t.Errorf("Error() = %v, want %v", tt.err.Error(), tt.wantMessage)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name        string
		err         *ValidationError
		wantMessage string
		targetError error
		shouldMatch bool
	}{
		{
			name: "invalid field",
			err: &ValidationError{
				Field:   "memory",
				Message: "must be a positive integer",
			},
			wantMessage: "validation failed for memory: must be a positive integer",
			targetError: ErrInvalidConfig,
			shouldMatch: true,
		},
		{
			name: "does not match other error",
			err: &ValidationError{
				Field:   "memory",
				Message: "invalid",
			},
			targetError: ErrNotFound,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantMessage != "" && tt.err.Error() != tt.wantMessage {
				t.Errorf("ValidationError.Error() = %v, want %v", tt.err.Error(), tt.wantMessage)
			}
			if errors.Is(tt.err, tt.targetError) != tt.shouldMatch {
				t.Errorf("errors.Is() match = %v, want %v", !tt.shouldMatch, tt.shouldMatch)
			}
		})
	}
}

func TestErrorHelpers(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		checks []struct {
			fn   func(error) bool
			want bool
			desc string
		}
	}{
		{
			name: "resource not found",
			err: &ResourceNotFoundError{
				ResourceType: "container",
				ID:           "abc123",
			},
			checks: []struct {
				fn   func(error) bool
				want bool
				desc string
			}{
				{IsNotFound, true, "IsNotFound"},
				{IsAlreadyExists, false, "IsAlreadyExists"},
				{IsInvalidConfig, false, "IsInvalidConfig"},
			},
		},
		{
			name: "resource exists",
			err: &ResourceExistsError{
				ResourceType: "container",
				ID:           "abc123",
			},
			checks: []struct {
				fn   func(error) bool
				want bool
				desc string
			}{
				{IsNotFound, false, "IsNotFound"},
				{IsAlreadyExists, true, "IsAlreadyExists"},
				{IsInvalidConfig, false, "IsInvalidConfig"},
			},
		},
		{
			name: "invalid config",
			err: &ConfigError{
				Field:   "port",
				Message: "invalid",
			},
			checks: []struct {
				fn   func(error) bool
				want bool
				desc string
			}{
				{IsNotFound, false, "IsNotFound"},
				{IsAlreadyExists, false, "IsAlreadyExists"},
				{IsInvalidConfig, true, "IsInvalidConfig"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, check := range tt.checks {
				if got := check.fn(tt.err); got != check.want {
					t.Errorf("%s() = %v, want %v", check.desc, got, check.want)
				}
			}
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	wrappedErr := Wrap(baseErr, "additional context")

	if wrappedErr.Error() != "additional context: base error" {
		t.Errorf("Wrap() error message = %v, want %v", wrappedErr.Error(), "additional context: base error")
	}

	if !errors.Is(wrappedErr, baseErr) {
		t.Error("Wrapped error should match base error with errors.Is")
	}
}

func TestNewError(t *testing.T) {
	message := "test error"
	err := New(message)

	if err.Error() != message {
		t.Errorf("New() error message = %v, want %v", err.Error(), message)
	}
}

func TestTimeoutAndCancellation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		fn   func(error) bool
		want bool
	}{
		{
			name: "timeout error",
			err:  ErrTimeout,
			fn:   IsTimeout,
			want: true,
		},
		{
			name: "canceled error",
			err:  ErrCanceled,
			fn:   IsCanceled,
			want: true,
		},
		{
			name: "timeout not canceled",
			err:  ErrTimeout,
			fn:   IsCanceled,
			want: false,
		},
		{
			name: "canceled not timeout",
			err:  ErrCanceled,
			fn:   IsTimeout,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(tt.err); got != tt.want {
				t.Errorf("Error check = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDaemonNotRunningError(t *testing.T) {
	tests := []struct {
		name        string
		err         *DaemonNotRunningError
		wantMessage string
		targetError error
		shouldMatch bool
	}{
		{
			name: "daemon not running",
			err: &DaemonNotRunningError{
				Message: "connection refused",
			},
			wantMessage: "docker daemon is not running: connection refused",
			targetError: ErrDaemonNotRunning,
			shouldMatch: true,
		},
		{
			name: "does not match other error",
			err: &DaemonNotRunningError{
				Message: "connection refused",
			},
			targetError: ErrNotFound,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantMessage != "" && tt.err.Error() != tt.wantMessage {
				t.Errorf("DaemonNotRunningError.Error() = %v, want %v", tt.err.Error(), tt.wantMessage)
			}
			if errors.Is(tt.err, tt.targetError) != tt.shouldMatch {
				t.Errorf("errors.Is() match = %v, want %v", !tt.shouldMatch, tt.shouldMatch)
			}
		})
	}
}
