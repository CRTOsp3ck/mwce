// internal/service/helper.go

package service

import (
	"fmt"
	"time"
)

// Helper function to format money
func formatMoney(amount int) string {
	if amount >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(amount)/1000000)
	} else if amount >= 1000 {
		return fmt.Sprintf("%.1fK", float64(amount)/1000)
	}
	return fmt.Sprintf("%d", amount)
}

// formatMessage formats a message with a title and content
func formatMessage(title string, format string, a ...interface{}) string {
	content := fmt.Sprintf(format, a...)
	return fmt.Sprintf("%s: %s", title, content)
}

// ptrTime creates a pointer to a time.Time
func ptrTime(t time.Time) *time.Time {
	return &t
}
