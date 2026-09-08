package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

var ErrAINotConfigured = errors.New("AI 助手尚未配置")

type OpenAIMessage struct {
	Role       string          `json:"role"`
	Content    string          `json:"content,omitempty"`
	ToolCalls  json.RawMessage `json:"tool_calls,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
}

type OpenAIToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type AIStreamResult struct {
	Content   string
	ToolCalls []OpenAIToolCall
}

type AIClient struct {
	baseURL, apiKey, model string
	client                 *http.Client
}

func NewAIClient(baseURL, apiKey, model string, connectTimeout, responseTimeout time.Duration) *AIClient {
	transport := &http.Transport{DialContext: (&net.Dialer{Timeout: connectTimeout}).DialContext}
	return &AIClient{baseURL: strings.TrimRight(baseURL, "/"), apiKey: apiKey, model: model, client: &http.Client{Transport: transport, Timeout: responseTimeout}}
}

func (c *AIClient) Stream(ctx context.Context, messages []OpenAIMessage, onDelta func(string) error) (*AIStreamResult, error) {
	if c.baseURL == "" {
		return nil, ErrAINotConfigured
	}
	tools := []any{map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        "recommend_seats",
			"description": "仅当用户给出明确日期、开始结束时间并请求选座时返回最多3个候选座位。seat_id 必须来自上下文。",
			"parameters": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"date", "start_time", "end_time", "candidates"},
				"properties": map[string]any{
					"date":       map[string]any{"type": "string"},
					"start_time": map[string]any{"type": "string"},
					"end_time":   map[string]any{"type": "string"},
					"candidates": map[string]any{
						"type": "array", "maxItems": 3,
						"items": map[string]any{
							"type": "object", "additionalProperties": false,
							"required": []string{"seat_id", "reason"},
							"properties": map[string]any{
								"seat_id": map[string]any{"type": "integer"},
								"reason":  map[string]any{"type": "string"},
							},
						},
					},
				},
			},
		},
	}}
	body, _ := json.Marshal(map[string]any{"model": c.model, "messages": messages, "tools": tools, "tool_choice": "auto", "stream": true})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("AI 服务返回 %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	result := &AIStreamResult{}
	calls := map[int]*OpenAIToolCall{}
	s := bufio.NewScanner(resp.Body)
	s.Buffer(make([]byte, 4096), 1024*1024)
	for s.Scan() {
		line := s.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content   string `json:"content"`
					ToolCalls []struct {
						Index    int    `json:"index"`
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if json.Unmarshal([]byte(data), &chunk) != nil || len(chunk.Choices) == 0 {
			continue
		}
		d := chunk.Choices[0].Delta
		if d.Content != "" {
			result.Content += d.Content
			if err := onDelta(d.Content); err != nil {
				return nil, err
			}
		}
		for _, part := range d.ToolCalls {
			tc := calls[part.Index]
			if tc == nil {
				tc = &OpenAIToolCall{}
				calls[part.Index] = tc
			}
			if part.ID != "" {
				tc.ID = part.ID
			}
			if part.Type != "" {
				tc.Type = part.Type
			}
			tc.Function.Name += part.Function.Name
			tc.Function.Arguments += part.Function.Arguments
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	for i := 0; i < len(calls); i++ {
		if calls[i] != nil {
			result.ToolCalls = append(result.ToolCalls, *calls[i])
		}
	}
	return result, nil
}
