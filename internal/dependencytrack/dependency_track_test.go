package dependencytrack

import (
	"net/http"
	"net/http/httptest"
)

// setup sets up a test HTTP server and a client configured to talk to it
func setup() (client *Client, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()

	server := httptest.NewServer(mux)

	client = New(WithAddress(server.URL))

	return client, mux, server.Close
}
