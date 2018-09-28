package util

import "time"

func OrString(newVal *string, oldVal string) string {
	if newVal != nil {
		return *newVal
	}

	return oldVal
}

func OrOptionalString(newVal, oldVal *string) *string {
	if newVal != nil {
		return newVal
	}

	return oldVal
}

func OrOptionalTime(newVal, oldVal *time.Time) *time.Time {
	if newVal != nil {
		return newVal
	}

	return oldVal
}
