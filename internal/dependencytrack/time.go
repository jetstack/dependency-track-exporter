package dependencytrack

import (
	"encoding/json"
	"time"
)

// Time is a custom time type that supports unmarshalling from a unix timestamp
type Time struct {
	time.Time
}

// UnmarshalJSON converts a unix timestamp to a time.Time
func (t *Time) UnmarshalJSON(data []byte) error {
	var v int64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	t.Time = time.Unix(v, 0)

	return nil
}
