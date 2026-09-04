package yourservice

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientRespectsMaxConcurrent(t *testing.T) {
	var peak atomic.Int32
	var active atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := active.Add(1)
		for {
			old := peak.Load()
			if cur <= old {
				break
			}
			if peak.CompareAndSwap(old, cur) {
				break
			}
		}
		time.Sleep(40 * time.Millisecond)
		active.Add(-1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client, err := New(srv.URL, "secret", &Options{
		MaxConcurrent: 3,
		RPS:           500,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	const workers = 25
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
			if err != nil {
				t.Errorf("NewRequest: %v", err)
				return
			}
			resp, err := client.HTTP.Do(req)
			if err != nil {
				t.Errorf("Do: %v", err)
				return
			}
			_ = resp.Body.Close()
		}()
	}
	wg.Wait()

	if got := peak.Load(); got > 3 {
		t.Fatalf("expected peak concurrent <= 3, got %d", got)
	}
}

func TestClientRateLimitsSerialRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	// Burst is derived from ceil(RPS)=2; first two requests are cheap, then ~500ms between tokens at 2 RPS.
	client, err := New(srv.URL, "secret", &Options{
		MaxConcurrent: 10,
		RPS:           2,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	start := time.Now()
	for i := 0; i < 6; i++ {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		resp, err := client.HTTP.Do(req)
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		_ = resp.Body.Close()
	}
	elapsed := time.Since(start)
	if elapsed < 1200*time.Millisecond {
		t.Fatalf("expected serial requests to take at least ~1.2s with RPS=2, got %v", elapsed)
	}
}

func TestNewRejectsInvalidLimits(t *testing.T) {
	_, err := New("https://api.example.com", "k", &Options{MaxConcurrent: -1, RPS: 1})
	if err == nil {
		t.Fatal("expected error for negative max concurrent")
	}
	_, err = New("https://api.example.com", "k", &Options{MaxConcurrent: 1, RPS: -0.5})
	if err == nil {
		t.Fatal("expected error for negative RPS")
	}
}
