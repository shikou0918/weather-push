package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/shikou0918/weather-push/request"
)

// ===== 共通ユーティリティ =====
func mustGetenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		fmt.Fprintf(os.Stderr, "missing required env: %s\n", key)
		os.Exit(1)
	}
	return val
}

// ===== LINE Messaging API (Push) =====
func pushLine(channelAccessToken, userID, text string) error {
	body := map[string]any{
		"to": userID,
		"messages": []map[string]any{
			{"type": "text", "text": text},
		},
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "https://api.line.me/v2/bot/message/push", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+channelAccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("line push http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("line push failed: status=%d", resp.StatusCode)
	}
	return nil
}

func main() {
	// 必須環境変数
	appID := mustGetenv("YAHOO_CLIENT_ID") // Yahoo!デベロッパーネットワークのアプリケーションID
	lat := mustGetenv("LAT")               // 例: 35.681236
	lon := mustGetenv("LON")               // 例: 139.767125
	lineToken := mustGetenv("LINE_CHANNEL_ACCESS_TOKEN")
	lineUserID := mustGetenv("LINE_USER_ID")

	// 任意（通知の見出し用）
	place := os.Getenv("PLACE_NAME")
	if place == "" {
		place = "指定地点"
	}

	// Yahoo API から情報を取得
	lines, err := request.FetchYahooWeather(appID, lat, lon)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// メッセージ組み立て
	msg := fmt.Sprintf("【%sの情報（観測/予報）】\n%s\n（毎朝7:00配信）",
		place,
		func() string {
			max := 7
			if len(lines) < max {
				max = len(lines)
			}
			return "- " + joinWithPrefix(lines[:max], "\n- ")
		}(),
	)

	// LINE Push
	if err := pushLine(lineToken, lineUserID, msg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("OK: pushed.")
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
