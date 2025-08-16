package request

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// 個別の天気データ
type WeatherData struct {
	Type string `json:"Type"`
	Date string `json:"Date"`
}

// WeatherList
type WeatherList struct {
	Weather []WeatherData `json:"Weather"`
}

// Property
type Property struct {
	WeatherList WeatherList `json:"WeatherList"`
}

// Feature
type Feature struct {
	Property Property `json:"Property"`
}

// Yahoo! JAPAN 気象情報 API のトップレベルレスポンス
type YahooWeatherResponse struct {
	Feature []Feature `json:"Feature"`
}

func FetchYahooWeather(appID, lat, lon string) ([]string, error) {
	base := "https://map.yahooapis.jp/weather/V1/place"
	params := url.Values{}
	params.Set("coordinates", fmt.Sprintf("%s,%s", lon, lat)) // 注意：lon,lat の順
	params.Set("appid", appID)
	params.Set("output", "json")

	reqURL := base + "?" + params.Encode()
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("yahoo weather http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("yahoo weather status=%d", resp.StatusCode)
	}

	var yr YahooWeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&yr); err != nil {
		return nil, fmt.Errorf("yahoo weather decode error: %w", err)
	}
	fmt.Println(yr)

	if len(yr.Feature) == 0 {
		return nil, fmt.Errorf("yahoo weather: empty feature")
	}
	list := yr.Feature[0].Property.WeatherList.Weather
	if len(list) == 0 {
		return nil, fmt.Errorf("yahoo weather: empty list")
	}

	// 表示用に「時刻 (observation/forecast)」を組み立て
	var lines []string
	maxItems := 7 // 現在＋先60分（10分刻みで最大7件程度）を目安に
	for i, w := range list {
		if i >= maxItems {
			break
		}
		ts := w.Date
		pretty := ts
		if len(ts) == 12 {
			// 日本時間として扱う（APIは日本ローカル想定）
			pretty = fmt.Sprintf("%s/%s %s:%s",
				ts[4:6], ts[6:8], ts[8:10], ts[10:12],
			)
		}
		lines = append(lines, fmt.Sprintf("%s (%s)", pretty, w.Type))
	}

	return lines, nil
}
