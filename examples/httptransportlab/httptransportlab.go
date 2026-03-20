package httptransportlab

import (
	"context"
	"io"
	"net/http"
	"net/http/httptrace"
	"time"
)

type FetchOptions struct {
	DrainBody bool
	ReadLimit int64
}

type Observation struct {
	ConnID     string
	Reused     bool
	WasIdle    bool
	Dialed     bool
	StatusCode int
	BytesRead  int64
}

func NewTransport() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DisableCompression = true
	transport.MaxIdleConns = 64
	transport.MaxIdleConnsPerHost = 16
	transport.MaxConnsPerHost = 32
	transport.IdleConnTimeout = 90 * time.Second
	transport.ResponseHeaderTimeout = 0
	return transport
}

func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: NewTransport(),
		Timeout:   timeout,
	}
}

func Fetch(ctx context.Context, client *http.Client, url string, opts FetchOptions) (Observation, error) {
	var obs Observation

	trace := &httptrace.ClientTrace{
		ConnectStart: func(_, _ string) {
			obs.Dialed = true
		},
		GotConn: func(info httptrace.GotConnInfo) {
			obs.Reused = info.Reused
			obs.WasIdle = info.WasIdle
		},
	}

	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), http.MethodGet, url, nil)
	if err != nil {
		return obs, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return obs, err
	}
	defer resp.Body.Close()

	obs.ConnID = resp.Header.Get("X-Conn-ID")
	obs.StatusCode = resp.StatusCode

	if opts.DrainBody {
		read, err := io.Copy(io.Discard, resp.Body)
		obs.BytesRead = read
		return obs, err
	}

	limit := opts.ReadLimit
	if limit <= 0 {
		limit = 1
	}
	read, err := io.CopyN(io.Discard, resp.Body, limit)
	if err != nil && err != io.EOF {
		return obs, err
	}
	obs.BytesRead = read
	return obs, nil
}
