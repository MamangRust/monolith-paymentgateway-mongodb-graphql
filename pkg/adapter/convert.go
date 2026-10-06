package adapter

import "time"

// parseTime converts an RFC3339 timestamp string from a gRPC response into a
// *time.Time. Empty or malformed values become nil.
func parseTime(ts string) *time.Time {
	if ts == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return nil
	}
	return &t
}

// parseDate converts an RFC3339 timestamp string into a time.Time. Empty or
// malformed values become the zero time.
func parseDate(ts string) time.Time {
	if ts == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Time{}
	}
	return t
}
