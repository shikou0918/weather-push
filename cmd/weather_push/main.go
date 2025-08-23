package main

import (
	"fmt"
	"os"

	"github.com/shikou0918/weather-push/request"
)

const (
	TokyoAreaCode   = "130000"
	SaitamaAreaCode = "110000"
)

func mustGetenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		fmt.Fprintf(os.Stderr, "missing required env: %s\n", key)
		os.Exit(1)
	}
	return val
}

func main() {
	lineToken := mustGetenv("LINE_CHANNEL_ACCESS_TOKEN")
	lineUserID := mustGetenv("LINE_USER_ID")

	// 気象庁ホームページAPIから情報を取得
	lines, err := request.FetchJMAWeather(TokyoAreaCode)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
		fmt.Println("OK: pushed.", lines)

	// メッセージ組み立て
	msg := fmt.Sprintf("【%sの情報（観測/予報）】\n%s\n（毎朝7:00配信）",
		func() string {
			max := 7
			if len(lines) < max {
				max = len(lines)
			}
			return "- " + joinWithPrefix(lines[:max], "\n- ")
		}(),
	)

	// LINE Push
	if err := request.PushLine(lineToken, lineUserID, msg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// ユーティリティ：先頭にプレフィクスを入れて結合
func joinWithPrefix(arr []string, sep string) string {
	out := ""
	for i, s := range arr {
		if i == 0 {
			out += s
		} else {
			out += sep + s
		}
	}
	return out
}
