package dependencytrack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	// ErrNotFound is a not found error
	ErrNotFound = errors.New("not found")
)

// Client interacts with Dependency-Track via the API
type Client struct {
	c    *http.Client
	opts *options
}

// New creates a new client
func New(opts ...Option) *Client {
	return &Client{
		c:    &http.Client{},
		opts: makeOptions(opts...),
	}
}

func (c *Client) newRequest(method, path string, headers map[string]string, body interface{}) (*http.Request, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.opts.Address, path), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Api-Key", c.opts.APIKey)
	req.Header.Add("Content-Type", "application/json")
	if headers != nil {
		for key, value := range headers {
			req.Header.Add(key, value)

		}
	}

	return req, nil
}

func (c *Client) do(req *http.Request, data interface{}) error {
	resp, err := c.c.Do(req)
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("Couldn't retrieve response from request")
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()
	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("error: %s: %d", req.URL.String(), resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &data); err != nil {
			return err
		}
	}

	return nil
}
