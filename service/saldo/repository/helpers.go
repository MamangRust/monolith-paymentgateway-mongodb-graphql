package repository

import "time"

// amountToInt64 widens a request amount to the 64-bit representation stored in
// MongoDB.
func amountToInt64(amount int) int64 {
	return int64(amount)
}

// int32Ptr narrows an optional 64-bit amount to the proto-friendly 32-bit
// representation.
func int32Ptr(v *int64) *int32 {
	if v == nil {
		return nil
	}
	result := int32(*v)
	return &result
}

// derefTime returns the zero time when the pointer is nil.
func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
