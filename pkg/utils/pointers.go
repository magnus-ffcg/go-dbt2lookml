package utils

// StringPtr returns a pointer to the given string value.
// This is a convenience function to avoid creating intermediate variables.
func StringPtr(s string) *string {
	return &s
}

// BoolPtr returns a pointer to the given bool value.
// This is a convenience function to avoid creating intermediate variables.
func BoolPtr(b bool) *bool {
	return &b
}

// IntPtr returns a pointer to the given int value.
// This is a convenience function to avoid creating intermediate variables.
func IntPtr(i int) *int {
	return &i
}

// Int64Ptr returns a pointer to the given int64 value.
// This is a convenience function to avoid creating intermediate variables.
func Int64Ptr(i int64) *int64 {
	return &i
}

// Float64Ptr returns a pointer to the given float64 value.
// This is a convenience function to avoid creating intermediate variables.
func Float64Ptr(f float64) *float64 {
	return &f
}
