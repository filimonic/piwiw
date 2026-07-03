package piwiw

import (
	"fmt"
	"log"
)

func logI(requestID, format string, args ...any) {
	logEvent(requestID, "I", format, args...)
}

func logW(requestID, format string, args ...any) {
	logEvent(requestID, "W", format, args...)
}

func logE(requestID, format string, args ...any) {
	logEvent(requestID, "E", format, args...)
}

func logEvent(requestID, level, format string, args ...any) {
	log.Printf("[%s] %s %s", requestID, level, fmt.Sprintf(format, args...))
}
