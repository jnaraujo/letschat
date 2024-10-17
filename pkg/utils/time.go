package utils

import (
	"fmt"
	"time"
)

func FormatDuration(d time.Duration) string {
	if d.Hours() >= 1 {
		hours := int(d.Hours())
		return fmt.Sprintf("%d hour%s ago", hours, Plural(hours))
	} else if d.Minutes() >= 1 {
		minutes := int(d.Minutes())
		return fmt.Sprintf("%d minute%s ago", minutes, Plural(minutes))
	} else {
		seconds := int(d.Seconds())
		return fmt.Sprintf("%d second%s ago", seconds, Plural(seconds))
	}
}
