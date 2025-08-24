package request

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	TokyoArea = "東京地方" // 天気/降水確率のエリア名
	TokyoCity = "東京"     // 気温のエリア名（配列ではこちらに入る）
)

type ForecastResponse []struct {
	ReportDatetime string       `json:"reportDatetime"`
	TimeSeries     []TimeSeries `json:"timeSeries"`
}

type TimeSeries struct {
	TimeDefines []string `json:"timeDefines"`
	Areas       []Area   `json:"areas"`
}

type Area struct {
	AreaInfo     AreaInfo `json:"area"`
	Weathers     []string `json:"weathers,omitempty"`
	Pops []string `json:"pops,omitempty"`

	Temps         []string `json:"temps,omitempty"`
}

type AreaInfo struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

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

	var lines []string
	appendSection := func(title string, rows []string) {
		if len(rows) == 0 {
			return
		}
		lines = append(lines, "【"+title+"】")
		lines = append(lines, rows...)
	}

	// --- 短期ブロック（fr[0]）から必要なものを拾う ---
	tsList := fr[0].TimeSeries

	// 1) 天気（東京地方）: 単位なし
	if rows := extractTimeSeries(tsList, TokyoArea, func(a Area) ([]string, bool) { return a.Weathers, len(a.Weathers) > 0 }, ""); len(rows) > 0 {
		appendSection("天気（東京地方）", rows)
	}

	// 2) 降水確率（東京地方）: 単位 %
	if rows := extractTimeSeries(tsList, TokyoArea, func(a Area) ([]string, bool) { return a.Pops, len(a.Pops) > 0 }, "%"); len(rows) > 0 {
		appendSection("降水確率（東京地方）", rows)
	}

	// 3) 気温（東京）: 単位 
	if rows := extractTimeSeries(tsList, TokyoCity, func(a Area) ([]string, bool) { return a.Temps, len(a.Temps) > 0 }, "℃"); len(rows) > 0 {
		appendSection("気温（東京）", rows)
	}

	// ※ 週間の tempsMin/Max 等は fr[1] 側。必要になったら同じ仕組みで追加可能。

	return lines, nil
}

// 指定したエリア名の時系列から、値スライスを取り出して「時刻: 値(+unit)」に整形して返す。
// getter は Area から対象フィールド（Weathers/Pops/Temps など）を取り出す関数。
// unit は "℃" や "%" を渡す。空なら付けない。
func extractTimeSeries(
	tsList []TimeSeries,
	areaName string,
	getter func(Area) ([]string, bool),
	unit string,
) []string {
	var out []string

	for _, ts := range tsList {
		// 該当エリアを探す
		idx := -1
		for i, a := range ts.Areas {
			if a.AreaInfo.Name == areaName {
				idx = i
				break
			}
		}
		if idx == -1 {
			continue
		}

		values, ok := getter(ts.Areas[idx])
		if !ok || len(values) == 0 {
			continue
		}

		// timeDefines と values の短い方に合わせる
		n := len(ts.TimeDefines)
		if len(values) < n {
			n = len(values)
		}

		for i := 0; i < n; i++ {
			val := values[i]
			if val == "" { // 空はスキップ（週間などで空が混じることがある）
				continue
			}
			if unit != "" {
				val = val + unit
			}

			tstr := ts.TimeDefines[i]
			if parsed, err := time.Parse(time.RFC3339, tstr); err == nil {
				// ここでは頭に "- " は付けない（main 側で「・」を付ける）
				out = append(out, fmt.Sprintf("%s: %s", parsed.Format("2006/01/02 (Mon) 15:04"), val))
			} else {
				out = append(out, fmt.Sprintf("%s: %s", tstr, val))
			}
		}
	}

	return out
}
