package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/SurveyController/SurveyConsole/internal/models"
)

const answerDatetimeWindowLayout = "2006-01-02 15:04:05"

// NormalizeAnswerDatetimeWindow keeps only datetime strings accepted by the
// original Python runtime format: YYYY-MM-DD HH:MM:SS.
func NormalizeAnswerDatetimeWindow(value [2]string) [2]string {
	return [2]string{
		formatAnswerDatetime(parseAnswerDatetime(value[0])),
		formatAnswerDatetime(parseAnswerDatetime(value[1])),
	}
}

func buildAnswerDatetimeWindowMS(cfg *models.RuntimeConfig) ([2]int64, error) {
	var zero [2]int64
	if cfg == nil || !supportsAnswerDatetimeWindow(cfg.SurveyProvider) {
		return zero, nil
	}

	window := NormalizeAnswerDatetimeWindow(cfg.AnswerDatetimeWindow)
	startText, endText := window[0], window[1]
	if startText == "" && endText == "" {
		return zero, nil
	}
	if startText == "" || endText == "" {
		return zero, fmt.Errorf("见数作答时间窗未配完整，请先设置开始和结束日期时间")
	}

	start := parseAnswerDatetime(startText)
	end := parseAnswerDatetime(endText)
	if start == nil || end == nil {
		return zero, fmt.Errorf("见数作答时间窗格式无效，请使用 YYYY-MM-DD HH:MM:SS")
	}
	if !end.After(*start) {
		return zero, fmt.Errorf("见数结束日期时间必须晚于开始日期时间")
	}
	maxDurationSeconds := maxInt(0, cfg.AnswerDuration[1])
	if int(end.Sub(*start).Seconds()) < maxDurationSeconds {
		return zero, fmt.Errorf("见数作答时间窗太窄，容不下当前最长作答时长")
	}
	return [2]int64{start.UnixMilli(), end.UnixMilli()}, nil
}

func supportsAnswerDatetimeWindow(provider string) bool {
	return strings.EqualFold(strings.TrimSpace(provider), models.ProviderCredamo)
}

func parseAnswerDatetime(value string) *time.Time {
	text := strings.TrimSpace(value)
	if text == "" {
		return nil
	}
	parsed, err := time.ParseInLocation(answerDatetimeWindowLayout, text, time.Local)
	if err != nil {
		return nil
	}
	return &parsed
}

func formatAnswerDatetime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(answerDatetimeWindowLayout)
}
