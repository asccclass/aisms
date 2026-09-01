package datefmt

import (
	"fmt"
	"strings"
	"time"
)

const DateLayout = "2006-01-02"

func NormalizeDate(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return ""
	}
	if len(text) == 10 && strings.Count(text, "-") == 2 {
		return text
	}
	if len(text) == 8 && !strings.ContainsAny(text, "-/") {
		return fmt.Sprintf("%s-%s-%s", text[:4], text[4:6], text[6:8])
	}
	if len(text) == 10 && strings.Count(text, "/") == 2 {
		return strings.ReplaceAll(text, "/", "-")
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
		"2006/1/2",
		"2006-1-2",
		DateLayout,
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, text); err == nil {
			return t.Format(DateLayout)
		}
	}
	return text
}

func Today() string {
	return time.Now().Format(DateLayout)
}
