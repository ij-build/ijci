package resource

import "time"

func orString(newVal *string, oldVal string) string {
	if newVal != nil {
		return *newVal
	}

	return oldVal
}

func orOptionalString(newVal, oldVal *string) *string {
	if newVal != nil {
		return newVal
	}

	return oldVal
}

func orOptionalTime(newVal, oldVal *time.Time) *time.Time {
	if newVal != nil {
		return newVal
	}

	return oldVal
}
