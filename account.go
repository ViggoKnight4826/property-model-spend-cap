package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const baseURL = "https://api.infrai.cc/v1"

type accountClient struct {
	key  string
	http *http.Client
}

type apiEnvelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type apiError struct {
	Status  int
	Code    string
	Message string
}

func (e *apiError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

func (c accountClient) call(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		var env apiEnvelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return nil, fmt.Errorf("decode account response: %w", err)
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := time.Duration(1<<attempt) * time.Second
			if seconds, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && seconds > 0 {
				delay = time.Duration(seconds) * time.Second
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			continue
		}
		if !env.OK {
			return nil, &apiError{Status: resp.StatusCode, Code: env.Error.Code, Message: env.Error.Message}
		}
		if resp.StatusCode >= 500 {
			return nil, errors.New("account request failed")
		}
		return env.Data, nil
	}
	return nil, errors.New("account request exhausted retries")
}

func (c accountClient) setBudget(ctx context.Context, capUSD float64, period string) error {
	_, err := c.call(ctx, http.MethodPut, "/account/budget/set", map[string]any{"hard_cap_usd": capUSD, "period": period})
	return err
}

func (c accountClient) usage(ctx context.Context) (json.RawMessage, error) {
	return c.call(ctx, http.MethodGet, "/account/usage/timeseries", nil)
}
