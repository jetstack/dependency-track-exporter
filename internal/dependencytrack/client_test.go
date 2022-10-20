package dependencytrack

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestClientRequest(t *testing.T) {
	type testSchema struct {
		Name       string
		Parameters []string
	}
	expectedReqBody := &testSchema{
		Name:       "foo",
		Parameters: []string{"one", "two", "three"},
	}
	expectedRespBody := &testSchema{
		Name:       "bar",
		Parameters: []string{"apple", "orange", "banana"},
	}

	client, mux, teardown := setup()
	defer teardown()

	client.opts.APIKey = "FAKEAPIKEY"

	mux.HandleFunc("/foobar", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Api-Key"); !cmp.Equal(got, "FAKEAPIKEY") {
			t.Errorf("Got X-Api-Key header %v, wanted %v", got, "FAKEAPIKEY")
		}

		if got := r.Header.Get("Content-type"); !cmp.Equal(got, "application/json") {
			t.Errorf("Got Content-type header %v, wanted %v", got, "application/json")
		}

		got := &testSchema{}
		if err := json.NewDecoder(r.Body).Decode(got); err != nil {
			t.Errorf("Unexpected error decoding request body: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !cmp.Equal(got, expectedReqBody) {
			t.Errorf("Expected request body %v, got %v", expectedReqBody, got)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedRespBody)
	})

	req, err := client.newRequest(http.MethodPost, "/foobar", expectedReqBody)
	if err != nil {
		t.Fatal(err)
	}
	data := &testSchema{}
	if err := client.do(req, data); err != nil {
		t.Fatal(err)
	}
}
