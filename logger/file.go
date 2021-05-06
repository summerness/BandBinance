package logger

import (
"fmt"
"time"
)

// getLogFilePath get the log file save path
func getLogFilePath() string {
	return fmt.Sprintf("%s%s", "data/","logs/")
}

// getLogFileName get the save name of the log file
func getLogFileName() string {
	return fmt.Sprintf("%s%s.%s",
		"log",
		time.Now().Format("20060102"),
		"log",
	)
}

