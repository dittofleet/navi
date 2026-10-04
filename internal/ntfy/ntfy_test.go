package ntfy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParsePriority(t *testing.T) {
	for in, want := range map[string]int{"min": 1, "LOW": 2, "default": 3, "high": 4, "urgent": 5, "max": 5, "1": 1, " 5 ": 5} {
		got, err := ParsePriority(in)
		if err != nil || got != want {
			t.Errorf("ParsePriority(%q) = %d, %v, want %d", in, got, err, want)
		}
	}
	for _, in := range []string{"", "0", "6", "loud", "2.5"} {
		if _, err := ParsePriority(in); err == nil {
			t.Errorf("ParsePriority(%q) succeeded, want an error", in)
		}
	}
}

func TestPublishSendsJSONWithToken(t *testing.T) {
	var got Message
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		if r.Method != http.MethodPost || r.URL.Path != "/" {
			t.Errorf("got %s %s, want POST /", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	msg := Message{Topic: "t", Message: "done", Title: "lichen on devbox ✅", Priority: 4, Tags: []string{"tada"}}
	if err := Publish(context.Background(), srv.URL, "tk_secret", msg); err != nil {
		t.Fatal(err)
	}
	if auth != "Bearer tk_secret" {
		t.Errorf("Authorization = %q", auth)
	}
	if got.Topic != "t" || got.Message != "done" || got.Title != msg.Title || got.Priority != 4 || len(got.Tags) != 1 {
		t.Errorf("server got %+v", got)
	}
}

func TestPublishOmitsAuthWithoutToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a := r.Header.Get("Authorization"); a != "" {
			t.Errorf("Authorization = %q, want none", a)
		}
	}))
	defer srv.Close()
	if err := Publish(context.Background(), srv.URL, "", Message{Topic: "t", Message: "m"}); err != nil {
		t.Fatal(err)
	}
}

func TestPublishReportsRefusal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"code":40301,"http":403,"error":"forbidden"}`))
	}))
	defer srv.Close()
	err := Publish(context.Background(), srv.URL, "", Message{Topic: "t", Message: "m"})
	if err == nil || !strings.Contains(err.Error(), "HTTP 403, forbidden") {
		t.Fatalf("err = %v", err)
	}
}

func TestPublishReportsBareStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`<html>bad gateway</html>`))
	}))
	defer srv.Close()
	err := Publish(context.Background(), srv.URL, "", Message{Topic: "t", Message: "m"})
	if err == nil || !strings.HasSuffix(err.Error(), "HTTP 502") {
		t.Fatalf("err = %v", err)
	}
}
