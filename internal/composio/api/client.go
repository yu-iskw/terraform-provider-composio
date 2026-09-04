// Copyright 2026 yu-iskw
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/sync/semaphore"
)

const (
	APIVersionPath       = "/api/v3.1"
	DefaultEndpoint      = "https://backend.composio.dev"
	DefaultTimeout       = 30 * time.Second
	DefaultMaxConcurrent = int64(8)
	maxResponseBytes     = 1 << 20
	headerProjectAPIKey  = "x-api-key"
	headerOrgKey         = "x-org-" + "api-key"
)

type AuthScope int

const (
	ScopeProject AuthScope = iota
	ScopeOrganization
)

type Options struct {
	Endpoint      string
	ProjectAPIKey string
	OrgAPIKey     string
	MaxConcurrent int64
	Timeout       time.Duration
	UserAgent     string
	Transport     http.RoundTripper
}

type Client struct {
	endpoint      string
	projectAPIKey string
	orgAPIKey     string
	userAgent     string
	http          *http.Client
	sem           *semaphore.Weighted
}

func New(opts Options) (*Client, error) {
	projectKey := strings.TrimSpace(opts.ProjectAPIKey)
	orgKey := strings.TrimSpace(opts.OrgAPIKey)
	if projectKey == "" && orgKey == "" {
		return nil, fmt.Errorf("at least one of api_key or org_api_key must be set")
	}

	endpoint := strings.TrimSpace(opts.Endpoint)
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	if err := validateEndpoint(endpoint); err != nil {
		return nil, err
	}

	maxC := opts.MaxConcurrent
	if maxC == 0 {
		maxC = DefaultMaxConcurrent
	}
	if maxC < 1 {
		return nil, fmt.Errorf("max concurrent requests must be at least 1, got %d", maxC)
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	if timeout < 0 {
		return nil, fmt.Errorf("request timeout cannot be negative")
	}

	ua := strings.TrimSpace(opts.UserAgent)
	if ua == "" {
		ua = "terraform-provider-composio terraform-plugin-framework"
	}

	base := opts.Transport
	if base == nil {
		base = http.DefaultTransport
	}

	return &Client{
		endpoint:      strings.TrimRight(endpoint, "/"),
		projectAPIKey: projectKey,
		orgAPIKey:     orgKey,
		userAgent:     ua,
		sem:           semaphore.NewWeighted(maxC),
		http: &http.Client{
			Timeout:   timeout,
			Transport: base,
		},
	}, nil
}

func (c *Client) Endpoint() string {
	return c.endpoint
}

func (c *Client) HasProjectKey() bool {
	return c.projectAPIKey != ""
}

func (c *Client) HasOrgKey() bool {
	return c.orgAPIKey != ""
}

func (c *Client) Do(ctx context.Context, scope AuthScope, method, path string, body any, out any) error {
	if err := c.requireScope(scope); err != nil {
		return err
	}

	var payload []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request body: %w", err)
		}
		payload = b
	}

	retryable := isRetryableMethod(method)
	var lastErr error
	attempts := 1
	if retryable {
		attempts = maxAttempts
	}

	for attempt := 1; attempt <= attempts; attempt++ {
		if err := c.sem.Acquire(ctx, 1); err != nil {
			return err
		}
		respBody, status, hdr, err := c.roundTrip(ctx, scope, method, path, payload)
		c.sem.Release(1)
		if err != nil {
			lastErr = err
			if retryable && isRetryableTransport(err) && attempt < attempts {
				if waitErr := sleep(ctx, backoff(attempt, 0)); waitErr != nil {
					return waitErr
				}
				continue
			}
			return err
		}
		if status >= 200 && status < 300 {
			if out == nil || len(respBody) == 0 {
				return nil
			}
			if err := json.Unmarshal(respBody, out); err != nil {
				return fmt.Errorf("decode response for %s %s: %w", method, path, err)
			}
			return nil
		}
		apiErr := parseAPIError(method, path, status, hdr, respBody)
		lastErr = apiErr
		if retryable && isRetryableStatus(status) && attempt < attempts {
			if waitErr := sleep(ctx, backoff(attempt, apiErr.RetryAfter)); waitErr != nil {
				return waitErr
			}
			continue
		}
		return apiErr
	}
	return lastErr
}

func (c *Client) roundTrip(ctx context.Context, scope AuthScope, method, path string, payload []byte) ([]byte, int, http.Header, error) {
	u := c.endpoint + APIVersionPath + path
	var rdr io.Reader
	if payload != nil {
		rdr = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, rdr)
	if err != nil {
		return nil, 0, nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	switch scope {
	case ScopeOrganization:
		req.Header.Set(headerOrgKey, c.orgAPIKey)
	default:
		req.Header.Set(headerProjectAPIKey, c.projectAPIKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, nil, err
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, maxResponseBytes+1)
	b, err := io.ReadAll(limited)
	if err != nil {
		return nil, resp.StatusCode, resp.Header, err
	}
	if len(b) > maxResponseBytes {
		return nil, resp.StatusCode, resp.Header, fmt.Errorf("response body for %s %s exceeded %d bytes", method, path, maxResponseBytes)
	}
	return b, resp.StatusCode, resp.Header.Clone(), nil
}

func (c *Client) requireScope(scope AuthScope) error {
	switch scope {
	case ScopeOrganization:
		if c.orgAPIKey == "" {
			return fmt.Errorf("org_api_key is required for this operation")
		}
	default:
		if c.projectAPIKey == "" {
			return fmt.Errorf("api_key is required for this operation")
		}
	}
	return nil
}

func validateEndpoint(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("endpoint must be an http or https URL")
	}
	if u.Host == "" {
		return fmt.Errorf("endpoint must include a host")
	}
	return nil
}

func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
