package main

import (
	"fmt"
	"os"
	"time"

	"github.com/shikou0918/weather-push/request"
)

const TokyoAreaCode = "130000"

func main() {
	lineToken := mustGetenv("LINE_CHANNEL_ACCESS_TOKEN")
	lineUserID := mustGetenv("LINE_USER_ID")

	msg, err := request.ComposeJMAReport(TokyoAreaCode, time.Now())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := request.PushLine(lineToken, lineUserID, msg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func mustGetenv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "missing required env: %s\n", key)
		os.Exit(1)
	}
	return v
}
