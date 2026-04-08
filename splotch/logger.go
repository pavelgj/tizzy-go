package splotch

import (
	"os"
	"time"
)

func debugLog(msg string) {
	if os.Getenv("SPLOTCH_DEBUG") != "1" {
		return
	}
	f, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = f.WriteString(time.Now().Format("2006-01-02 15:04:05 ") + msg + "\n")
}
