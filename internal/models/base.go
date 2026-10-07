package models

import (
	"encoding/json"
	"time"

	"backend_go/pkg/format_errors"
)

type MessageResponse struct {
	Message string `json:"message"`
}

type Time time.Time

func (t *Time) ToTime() *time.Time {
	res := time.Time(*t)
	return &res
}

func (t *Time) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return format_errors.Wrap(err, "ошибка time")
	}
	var result time.Time
	if result, err = time.Parse("2006-01-02", s); err != nil {
		if result, err = time.Parse("2006-01-02 15:04:05", s); err != nil {
			if result, err = time.Parse("2006-01-02T15:04:05", s); err != nil {
				if result, err = time.Parse("2006-01-02T15:04:05-07:00", s); err != nil {
					if result, err = time.Parse(time.RFC3339, s); err != nil {
						return format_errors.Wrap(err, "ошибка time")
					}
				}
			}
		}
	}
	*t = Time(result)
	return nil
}
