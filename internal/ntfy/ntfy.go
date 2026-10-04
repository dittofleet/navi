// Package ntfy publishes messages to an ntfy server
// (https://docs.ntfy.sh/publish/).
package ntfy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Message is one notification. Only Topic and Message are required.
type Message struct {
	Topic    string   `json:"topic"`
	Message  string   `json:"message"`
	Title    string   `json:"title,omitempty"`
	Priority int      `json:"priority,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Click    string   `json:"click,omitempty"`
}

// priorityNames are ntfy's own words for its five levels. "max" is its
// alias for "urgent".
var priorityNames = map[string]int{
	"min": 1, "low": 2, "default": 3, "high": 4, "urgent": 5, "max": 5,
}

// ParsePriority accepts a level by name or by number, 1 through 5.
func ParsePriority(s string) (int, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if n, ok := priorityNames[s]; ok {
		return n, nil
	}
	if n, err := strconv.Atoi(s); err == nil && n >= 1 && n <= 5 {
		return n, nil
	}
	return 0, fmt.Errorf("priority must be 1-5 or min, low, default, high, urgent, got %q", s)
}

// Publish posts msg to server. It goes as JSON to the server root rather
// than as a body with header fields, because headers cannot carry
// non-ASCII text such as an emoji in a title.
func Publish(ctx context.Context, server, token string, msg Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, server, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not reach %s: %w", server, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 == 2 {
		return nil
	}

	// ntfy explains a refusal as {"code":..., "error":"..."}. Anything
	// else (a proxy's HTML page) is reduced to the status.
	var refusal struct {
		Error string `json:"error"`
	}
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if json.Unmarshal(data, &refusal) == nil && refusal.Error != "" {
		return fmt.Errorf("%s refused the message: HTTP %d, %s", server, resp.StatusCode, refusal.Error)
	}
	return fmt.Errorf("%s refused the message: HTTP %d", server, resp.StatusCode)
}
