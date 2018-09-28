package util

func JustStarted(oldStatus, newStatus string) bool {
	return newStatus == "in-progress" && oldStatus != "in-progress"
}

func JustCompleted(oldStatus, newStatus string) bool {
	return IsTerminal(newStatus) && !IsTerminal(oldStatus)
}

func IsTerminal(buildStatus string) bool {
	return buildStatus != "queued" && buildStatus != "in-progress"
}
