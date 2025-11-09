package auth

import (
	"fmt"
	"time"
)

func parseFlexibleTime(s string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,          // 2006-01-02T15:04:05Z07:00
		"2006-01-02 15:04:05", // 2006-01-02 15:04:05
		"2006-01-02",          // 2006-01-02
	}

	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time %q: %v", s, err)
}
