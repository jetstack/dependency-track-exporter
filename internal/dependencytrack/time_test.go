package dependencytrack

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// TestTime tests that a unix timestamp is unmarshalled into the expected
// time.Time
func TestTime(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	data := []byte(strconv.Itoa(int(now.Unix())))

	customTime := Time{}
	if err := json.Unmarshal(data, &customTime); err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(customTime.Time, now) {
		t.Errorf("Expected time %v but got %v", now, customTime.Time)
	}
}
