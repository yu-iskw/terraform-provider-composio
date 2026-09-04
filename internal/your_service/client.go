// Package yourservice holds the placeholder HTTP API client for the YOUR_SERVICE layer in this template provider.
// When you fork, rename the directory internal/your_service and this package to match your product API (see README.md here).
package yourservice

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
)

// DefaultEndpoint is used when the provider omits an explicit endpoint.
const DefaultEndpoint = "https://api.example.com"

// Defaults for rate and concurrency when the provider omits optional attributes.
const (
	DefaultMaxConcurrent = int64(10)
	DefaultRPS           = 10.0
)

// Options configures HTTP rate limiting and concurrency. Zero values mean "use defaults".
type Options struct {
	MaxConcurrent int64   // max in-flight HTTP requests (default DefaultMaxConcurrent)
	RPS           float64 // sustained requests per second for the token bucket (default DefaultRPS)
}

// Client is the template API client. Use HTTP for REST calls; transport applies auth, rate, and concurrency limits.
type Client struct {
	Endpoint string
	HTTP     *http.Client
}

// New validates configuration and returns a client for use as provider ResourceData/DataSourceData.
// opts may be nil; zero fields in opts select DefaultMaxConcurrent and DefaultRPS.
func New(endpoint, apiKey string, opts *Options) (*Client, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("api_key must not be empty")
	}

	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}

	maxC, rps := DefaultMaxConcurrent, DefaultRPS
	if opts != nil {
		if opts.MaxConcurrent < 0 {
			return nil, fmt.Errorf("max concurrent requests cannot be negative")
		}
		if opts.RPS < 0 {
			return nil, fmt.Errorf("requests per second cannot be negative")
		}
		if opts.MaxConcurrent > 0 {
			maxC = opts.MaxConcurrent
		}
		if opts.RPS > 0 {
			rps = opts.RPS
		}
	}
	if maxC < 1 {
		return nil, fmt.Errorf("max concurrent requests must be at least 1, got %d", maxC)
	}
	if rps <= 0 {
		return nil, fmt.Errorf("requests per second must be greater than 0, got %v", rps)
	}

	burst := int(math.Ceil(rps))
	if burst < 1 {
		burst = 1
	}
	if burst > 100 {
		burst = 100
	}

	lim := rate.NewLimiter(rate.Limit(rps), burst)
	sem := semaphore.NewWeighted(maxC)

	rt := &roundTripper{
		apiKey: apiKey,
		lim:    lim,
		sem:    sem,
		base:   http.DefaultTransport,
	}

	return &Client{
		Endpoint: endpoint,
		HTTP: &http.Client{
			Timeout:   60 * time.Second,
			Transport: rt,
		},
	}, nil
}

// roundTripper applies rate limit (wait before acquiring concurrency), API key auth, then delegates.
type roundTripper struct {
	apiKey string
	lim    *rate.Limiter
	sem    *semaphore.Weighted
	base   http.RoundTripper
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if err := rt.lim.Wait(ctx); err != nil {
		return nil, err
	}
	if err := rt.sem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer rt.sem.Release(1)

	r2 := req.Clone(ctx)
	if rt.apiKey != "" {
		r2.Header.Set("Authorization", "Bearer "+rt.apiKey)
	}
	return rt.base.RoundTrip(r2)
}
