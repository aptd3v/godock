package errdefs

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("resource not found")
	// ErrAlreadyExists is returned when attempting to create a resource that already exists
	ErrAlreadyExists = errors.New("resource already exists")
	// ErrInvalidConfig is returned when a configuration is invalid
	ErrInvalidConfig = errors.New("invalid configuration")
	// ErrDaemonNotRunning is returned when the Docker daemon is not running
	ErrDaemonNotRunning = errors.New("docker daemon is not running")
	// ErrTimeout is returned when an operation times out
	ErrTimeout = errors.New("operation timed out")
	// ErrCanceled is returned when an operation is canceled
	ErrCanceled = errors.New("operation canceled")
)

// ResourceNotFoundError represents a not found error for a specific resource
type ResourceNotFoundError struct {
	ResourceType string
	ID           string
}

func (e *ResourceNotFoundError) Error() string {

	return fmt.Sprintf("%s not found: %s", e.ResourceType, e.ID)
}

// Is implements the errors.Is interface
func (e *ResourceNotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

// ResourceExistsError represents an already exists error for a specific resource
type ResourceExistsError struct {
	ResourceType string
	ID           string
}

func (e *ResourceExistsError) Error() string {
	return fmt.Sprintf("%s already exists: %s", e.ResourceType, e.ID)
}

// Is implements the errors.Is interface
func (e *ResourceExistsError) Is(target error) bool {
	return target == ErrAlreadyExists
}

// ConfigError represents an invalid configuration error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("invalid configuration for %s: %s", e.Field, e.Message)
}

// Is implements the errors.Is interface
func (e *ConfigError) Is(target error) bool {
	return target == ErrInvalidConfig
}

// ContainerError represents a container-specific error
type ContainerError struct {
	ID      string
	Op      string
	Message string
}

func (e *ContainerError) Error() string {
	return fmt.Sprintf("container %s: %s failed: %s", e.ID, e.Op, e.Message)
}

// NetworkError represents a network-specific error
type NetworkError struct {
	ID      string
	Op      string
	Message string
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network %s: %s failed: %s", e.ID, e.Op, e.Message)
}

// VolumeError represents a volume-specific error
type VolumeError struct {
	Name    string
	Op      string
	Message string
}

func (e *VolumeError) Error() string {
	return fmt.Sprintf("volume %s: %s failed: %s", e.Name, e.Op, e.Message)
}

// ImageError represents an image-specific error
type ImageError struct {
	Ref     string
	Op      string
	Message string
}

func (e *ImageError) Error() string {
	return fmt.Sprintf("image %s: %s failed: %s", e.Ref, e.Op, e.Message)
}

// ExecError represents an exec-specific error
type ExecError struct {
	ID      string
	Op      string
	Message string
}

func (e *ExecError) Error() string {
	return fmt.Sprintf("exec %s: %s failed: %s", e.ID, e.Op, e.Message)
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// Is implements the errors.Is interface
func (e *ValidationError) Is(target error) bool {
	return target == ErrInvalidConfig
}

// DaemonNotRunningError represents an error when the Docker daemon is not running
type DaemonNotRunningError struct {
	Message string
}

func (e *DaemonNotRunningError) Error() string {
	return fmt.Sprintf("docker daemon is not running: %s", e.Message)
}

// Is implements the errors.Is interface
func (e *DaemonNotRunningError) Is(target error) bool {
	return target == ErrDaemonNotRunning
}

// New creates a new error with the given message
func New(message string) error {
	return errors.New(message)
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

// IsNotFound returns true if the error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists returns true if the error is an already exists error
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsInvalidConfig returns true if the error is an invalid config error
func IsInvalidConfig(err error) bool {
	return errors.Is(err, ErrInvalidConfig)
}

// IsDaemonNotRunning returns true if the error is a daemon not running error
func IsDaemonNotRunning(err error) bool {
	return errors.Is(err, ErrDaemonNotRunning)
}

// IsTimeout returns true if the error is a timeout error
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsCanceled returns true if the error is a canceled error
func IsCanceled(err error) bool {
	return errors.Is(err, ErrCanceled)
}
