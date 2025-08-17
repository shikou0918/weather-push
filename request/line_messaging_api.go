package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func PushLine(channelAccessToken, userID, text string) error {
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
