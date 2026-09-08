package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAIClientStreamsTextAndToolCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, `data: {"choices":[{"delta":{"content":"好的"}}]}`)
		fmt.Fprintln(w)
		fmt.Fprintln(w, `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"recommend_seats","arguments":"{\"date\":\"2026-09-09\"}"}}]}}]}`)
		fmt.Fprintln(w)
		fmt.Fprintln(w, "data: [DONE]")
	}))
	defer server.Close()
	client := NewAIClient(server.URL, "", "test", time.Second, time.Second)
	var streamed string
	result, err := client.Stream(context.Background(), []OpenAIMessage{{Role: "user", Content: "选座"}}, func(v string) error { streamed += v; return nil })
	if err != nil {
		t.Fatal(err)
	}
	if streamed != "好的" || result.Content != "好的" {
		t.Fatalf("unexpected content %q / %q", streamed, result.Content)
	}
	if len(result.ToolCalls) != 1 || result.ToolCalls[0].Function.Name != "recommend_seats" {
		t.Fatalf("unexpected calls %#v", result.ToolCalls)
	}
}

func TestAIClientRequiresConfiguration(t *testing.T) {
	client := NewAIClient("", "", "test", time.Second, time.Second)
	if _, err := client.Stream(context.Background(), nil, func(string) error { return nil }); err != ErrAINotConfigured {
		t.Fatalf("got %v", err)
	}
}
