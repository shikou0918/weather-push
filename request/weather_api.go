package request

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	TokyoArea = "東京地方"
	TokyoCity = "東京"
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
	AreaInfo AreaInfo `json:"area"`
	Weathers []string `json:"weathers,omitempty"`
	Pops     []string `json:"pops,omitempty"`
	Temps    []string `json:"temps,omitempty"`
}

type AreaInfo struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

func ComposeJMAReport(areaCode string, now time.Time) (string, error) {
	// 各行（見出し + 項目行）を取得
	lines, err := FetchJMAWeather(areaCode)
	if err != nil {
		return "", err
	}

	body := formatWithSectionBreaks(lines)

	title := "東京地方の天気・降水確率・気温（観測/予報）"
	msg := fmt.Sprintf(
		"【%s】\n\n%s\n\n（毎朝7:00配信 / %s）",
		title,
		body,
		now.Format("2006/01/02 15:04"),
	)
	return msg, nil
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

	// 短期の気象情報を取得
	tsList := fr[0].TimeSeries

	// 天気（東京地方）
	if rows := extractTimeSeries(tsList, TokyoArea, func(a Area) ([]string, bool) { return a.Weathers, len(a.Weathers) > 0 }, ""); len(rows) > 0 {
		appendSection("天気（東京地方）", rows)
	}
	// 降水確率（東京地方）
	if rows := extractTimeSeries(tsList, TokyoArea, func(a Area) ([]string, bool) { return a.Pops, len(a.Pops) > 0 }, "%"); len(rows) > 0 {
		appendSection("降水確率（東京地方）", rows)
	}
	// 気温（東京）
	if rows := extractTimeSeries(tsList, TokyoCity, func(a Area) ([]string, bool) { return a.Temps, len(a.Temps) > 0 }, "℃"); len(rows) > 0 {
		appendSection("気温（東京）", rows)
	}

	return lines, nil
}

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

		n := len(ts.TimeDefines)
		if len(values) < n {
			n = len(values)
		}

		for i := 0; i < n; i++ {
			val := values[i]
			if val == "" {
				continue
			}
			if unit != "" {
				val += unit
			}
			tstr := ts.TimeDefines[i]
			if parsed, err := time.Parse(time.RFC3339, tstr); err == nil {
				out = append(out, fmt.Sprintf("%s: %s", parsed.Format("2006/01/02 (Mon) 15:04"), val))
			} else {
				out = append(out, fmt.Sprintf("%s: %s", tstr, val))
			}
		}
	}
	return out
}

func formatWithSectionBreaks(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	var out []string
	first := true
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		if strings.HasPrefix(l, "【") {
			if !first {
				out = append(out, "")
			}
			out = append(out, l)
			first = false
		} else {
			out = append(out, "・"+l)
		}
	}
	return strings.Join(out, "\n")
}
