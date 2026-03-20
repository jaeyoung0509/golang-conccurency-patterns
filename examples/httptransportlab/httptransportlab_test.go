package httptransportlab

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestFetchReusesConnectionWhenBodyDrained(t *testing.T) {
	server := newTrackedServer(8<<10, 0)
	defer server.Close()

	client := NewClient(0)
	defer closeIdleConnections(client)

	first, err := Fetch(context.Background(), client, server.URL+"/reusable", FetchOptions{DrainBody: true})
	if err != nil {
		t.Fatalf("Fetch first returned error: %v", err)
	}

	second, err := Fetch(context.Background(), client, server.URL+"/reusable", FetchOptions{DrainBody: true})
	if err != nil {
		t.Fatalf("Fetch second returned error: %v", err)
	}

	if !second.Reused {
		t.Fatalf("second request reused = false, want true")
	}
	if first.ConnID != second.ConnID {
		t.Fatalf("connection IDs = %q and %q, want same connection", first.ConnID, second.ConnID)
	}
}

func TestFetchLosesReuseWhenBodyIsNotDrained(t *testing.T) {
	server := newTrackedServer(1<<20, 0)
	defer server.Close()

	client := NewClient(0)
	defer closeIdleConnections(client)

	first, err := Fetch(context.Background(), client, server.URL+"/large", FetchOptions{DrainBody: false, ReadLimit: 16})
	if err != nil {
		t.Fatalf("Fetch first returned error: %v", err)
	}
	second, err := Fetch(context.Background(), client, server.URL+"/large", FetchOptions{DrainBody: true})
	if err != nil {
		t.Fatalf("Fetch second returned error: %v", err)
	}

	if second.Reused {
		t.Fatalf("second request reused = true, want false after undrained body")
	}
	if first.ConnID == second.ConnID {
		t.Fatalf("connection IDs = %q and %q, want a new connection", first.ConnID, second.ConnID)
	}
}

func TestFetchHonorsContextCancellationWithoutPoisoningClient(t *testing.T) {
	server := newTrackedServer(4<<10, 120*time.Millisecond)
	defer server.Close()

	client := NewClient(0)
	defer closeIdleConnections(client)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := Fetch(ctx, client, server.URL+"/slow", FetchOptions{DrainBody: true})
	if err == nil {
		t.Fatalf("Fetch error = nil, want deadline exceeded")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Fetch error = %v, want context deadline exceeded", err)
	}
	if elapsed := time.Since(start); elapsed > 80*time.Millisecond {
		t.Fatalf("Fetch elapsed = %v, want fast cancellation", elapsed)
	}

	fast, err := Fetch(context.Background(), client, server.URL+"/fast", FetchOptions{DrainBody: true})
	if err != nil {
		t.Fatalf("follow-up Fetch returned error: %v", err)
	}
	if fast.StatusCode != http.StatusOK {
		t.Fatalf("follow-up status = %d, want 200", fast.StatusCode)
	}
}

func TestClientTimeoutBoundsWholeExchange(t *testing.T) {
	server := newTrackedServer(4<<10, 120*time.Millisecond)
	defer server.Close()

	client := NewClient(30 * time.Millisecond)
	defer closeIdleConnections(client)

	_, err := Fetch(context.Background(), client, server.URL+"/slow", FetchOptions{DrainBody: true})
	if err == nil {
		t.Fatalf("Fetch error = nil, want timeout")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !stringsContain(err.Error(), "Client.Timeout") {
		t.Fatalf("Fetch error = %v, want client timeout", err)
	}
}

type connIDKey struct{}

func newTrackedServer(bodySize int, headerDelay time.Duration) *httptest.Server {
	var nextConnID atomic.Int64

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if headerDelay > 0 && r.URL.Path == "/slow" {
			time.Sleep(headerDelay)
		}

		connID := r.Context().Value(connIDKey{}).(int64)
		w.Header().Set("X-Conn-ID", strconv.FormatInt(connID, 10))
		payload := bytes.Repeat([]byte("x"), bodySize)
		w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
		_, _ = w.Write(payload)
	}))

	server.Config.ConnContext = func(ctx context.Context, _ net.Conn) context.Context {
		id := nextConnID.Add(1)
		return context.WithValue(ctx, connIDKey{}, id)
	}

	server.Start()
	return server
}

func closeIdleConnections(client *http.Client) {
	if transport, ok := client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
}

func stringsContain(s, sub string) bool {
	return len(sub) == 0 || len(s) >= len(sub) && bytes.Contains([]byte(s), []byte(sub))
}
