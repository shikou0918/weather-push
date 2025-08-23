package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shikou0918/weather-push/request"
)

const TokyoAreaCode = "130000" // 東京地方コード

func main() {
	lineToken := mustGetenv("LINE_CHANNEL_ACCESS_TOKEN")
	lineUserID := mustGetenv("LINE_USER_ID")

	// 気象庁ホームページAPIから情報を取得（見出し行と項目行が混在して返る想定）
	lines, err := request.FetchJMAWeather(TokyoAreaCode)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 見出しの前に空行を入れ、項目行は「・」で箇条書きに整形
	body := formatWithSectionBreaks(lines)

	// メッセージ全体を組み立て（冒頭と各セクションの間も空ける）
	title := "東京地方の天気・降水確率・気温（観測/予報）"
	msg := fmt.Sprintf(
		"【%s】\n\n%s\n\n（毎朝7:00配信 / %s）",
		title,
		body,
		time.Now().Format("2006/01/02 15:04"),
	)

	// LINE Push
	if err := request.PushLine(lineToken, lineUserID, msg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// 見出し行（先頭が「【」）の前に空行を挿入し、項目行は「・」で箇条書きにする
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
			// 見出し：先頭以外なら空行を挟む
			if !first {
				out = append(out, "")
			}
			out = append(out, l)
			first = false
		} else {
			// 項目：先頭に「・」を付ける（二重ハイフン対策で余計なプレフィクスは付けない）
			out = append(out, "・"+l)
		}
	}
	return strings.Join(out, "\n")
}

func mustGetenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		fmt.Fprintf(os.Stderr, "missing required env: %s\n", key)
		os.Exit(1)
	}
	return val
}
