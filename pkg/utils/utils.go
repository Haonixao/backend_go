package utils

import (
	"fmt"
	"strings"
	"time"
)

var TimeLayouts = map[string]string{
	"YearOnly":                  "2006",
	"YearMonthOnly":             "2006-01",
	"DateOnly":                  "2006-01-02",
	"RussianYearMonthOnly":      "01.2006",
	"RussianDate":               "02.01.2006",
	"SlashYearMonthOnly":        "01/2006",
	"SlashDate":                 "02/01/2006",
	"ReverseSlashYearMonthOnly": "2006/01",
	"ReverseSlashDate":          "2006/01/02",

	"DateTime": "2006-01-02 15:04:05",
	"ISO8601":  "2006-01-02T15:04:05-07:00",
	"RFC3339":  time.RFC3339,
}

func ParseDate(date string) (time.Time, error) {
	for _, layout := range TimeLayouts {
		if result, err := time.Parse(layout, date); err == nil {
			return result, nil
		}
	}

	return time.Time{}, fmt.Errorf("failed parse date: %s", date)
}

func FormatTimeTrimmed(t time.Time, layoutName string) string {
	if layout, ok := TimeLayouts[layoutName]; ok {
		res := t.Format(layout)

		// Мы обрезаем только если компоненты времени "нулевые" (00 или 01 для дня/месяца)
		// И только если они находятся в конце строки.

		// Порядок обрезки: секунды -> минуты -> часы -> дни -> месяцы

		// 1. Секунды
		if t.Second() == 0 {
			res = strings.TrimSuffix(res, ":00")
			res = strings.TrimSuffix(res, ":00+00")
		} else {
			return res
		}

		// 2. Минуты
		if t.Minute() == 0 {
			res = strings.TrimSuffix(res, ":00")
		} else {
			return res
		}

		// 3. Часы
		if t.Hour() == 0 {
			res = strings.TrimSuffix(res, " 00")
			res = strings.TrimSuffix(res, "T00")
			res = strings.TrimSuffix(res, "T") // Для RFC3339 если все время нулевое
		} else {
			return res
		}

		// 4. Дни
		if t.Day() == 1 {
			res = strings.TrimSuffix(res, "-01")
			res = strings.TrimSuffix(res, ".01")
			res = strings.TrimSuffix(res, "/01")
			res = strings.TrimSuffix(res, " 01")
		} else {
			return res
		}

		// 5. Месяцы
		if t.Month() == 1 {
			res = strings.TrimSuffix(res, "-01")
			res = strings.TrimSuffix(res, ".01")
			res = strings.TrimSuffix(res, "/01")
			res = strings.TrimSuffix(res, " 01")
		}

		return res
	}
	return ""
}
