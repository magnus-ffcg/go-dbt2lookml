package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringPtr(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"simple string", "hello"},
		{"string with spaces", "hello world"},
		{"unicode string", "你好世界"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringPtr(tt.input)
			require.NotNil(t, result)
			assert.Equal(t, tt.input, *result)
		})
	}
}

func TestBoolPtr(t *testing.T) {
	tests := []struct {
		name  string
		input bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BoolPtr(tt.input)
			require.NotNil(t, result)
			assert.Equal(t, tt.input, *result)
		})
	}
}

func TestIntPtr(t *testing.T) {
	tests := []struct {
		name  string
		input int
	}{
		{"zero", 0},
		{"positive", 42},
		{"negative", -123},
		{"max int", int(^uint(0) >> 1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IntPtr(tt.input)
			require.NotNil(t, result)
			assert.Equal(t, tt.input, *result)
		})
	}
}

func TestInt64Ptr(t *testing.T) {
	tests := []struct {
		name  string
		input int64
	}{
		{"zero", 0},
		{"positive", 42},
		{"negative", -123},
		{"large number", 9223372036854775807},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Int64Ptr(tt.input)
			require.NotNil(t, result)
			assert.Equal(t, tt.input, *result)
		})
	}
}

func TestFloat64Ptr(t *testing.T) {
	tests := []struct {
		name  string
		input float64
	}{
		{"zero", 0.0},
		{"positive", 3.14},
		{"negative", -2.718},
		{"scientific", 1.23e-4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Float64Ptr(tt.input)
			require.NotNil(t, result)
			assert.Equal(t, tt.input, *result)
		})
	}
}

// TestPointerUniqueness verifies that each call creates a unique pointer
func TestPointerUniqueness(t *testing.T) {
	s1 := StringPtr("test")
	s2 := StringPtr("test")

	// Same value but different pointers
	assert.Equal(t, *s1, *s2)
	assert.NotSame(t, s1, s2, "Pointers should be different instances")

	// Modifying one doesn't affect the other
	*s1 = "modified"
	assert.Equal(t, "modified", *s1)
	assert.Equal(t, "test", *s2)
}
