package yourservice

import "testing"

func TestNewRequiresAPIKey(t *testing.T) {
	_, err := New("https://api.example.com", "", nil)
	if err == nil {
		t.Fatal("expected empty API key to return an error")
	}
}

func TestNewDefaultsEndpoint(t *testing.T) {
	client, err := New("", "test-key", nil)
	if err != nil {
		t.Fatalf("expected client creation to succeed: %v", err)
	}

	if client.Endpoint != DefaultEndpoint {
		t.Fatalf("expected endpoint %q, got %q", DefaultEndpoint, client.Endpoint)
	}
}
