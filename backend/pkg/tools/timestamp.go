package tools

import "time"

// Timestamp returns the current timestamp in a human-readable format("YYYY-MM-DD HH:MM:SS").
func Timestamp() string {
	timestamp := time.Now().Unix()
	formattedTime := time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
	return formattedTime
}
