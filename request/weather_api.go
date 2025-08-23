package request

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// 気象庁 API のレスポンス構造体（必要最低限に絞って定義）
type ForecastResponse []struct {
	PublishingOffice string `json:"publishingOffice"`
	ReportDatetime   string `json:"reportDatetime"`
	TimeSeries       []struct {
		TimeDefines []string `json:"timeDefines"`
		Areas       []struct {
			Area struct {
				Name string `json:"name"`
				Code string `json:"code"`
			} `json:"area"`
			Weathers []string `json:"weathers"`
		} `json:"areas"`
	} `json:"timeSeries"`
}

// FetchJMAWeather は気象庁のAPIから天気予報を取得する
func FetchJMAWeather(areaCode string) ([]string, error) {
	base := fmt.Sprintf("https://www.jma.go.jp/bosai/forecast/data/forecast/%s.json", areaCode)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(base)
	if err != nil {
		return nil, fmt.Errorf("jma weather http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("jma weather status=%d", resp.StatusCode)
	}

	var fr ForecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&fr); err != nil {
		return nil, fmt.Errorf("jma weather decode error: %w", err)
	}

	if len(fr) == 0 || len(fr[0].TimeSeries) == 0 {
		return nil, fmt.Errorf("jma weather: empty response")
	}

	// 例として最初のタイムシリーズ（天気予報）だけを抜き出す
	ts := fr[0].TimeSeries[0]
	var lines []string
	for i, t := range ts.TimeDefines {
		// 各エリアの天気をまとめる（とりあえず最初のエリア）
		if len(ts.Areas) > 0 && i < len(ts.Areas[0].Weathers) {
			weather := ts.Areas[0].Weathers[i]
			lines = append(lines, fmt.Sprintf("%s: %s", t, weather))
		}
	}

	return lines, nil
}
