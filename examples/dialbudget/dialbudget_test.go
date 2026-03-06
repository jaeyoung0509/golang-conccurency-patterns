package dialbudget

import (
	"context"
	"net"
	"net/netip"
	"testing"
	"time"
)

func TestProbeFirstHealthyFallsBackToReachableEndpoint(t *testing.T) {
	listener := startProbeServer(t, 0, "ok\n")
	defer listener.Close()

	endpoint := mustAddrPort(t, listener.Addr().String())
	unreachable := netip.MustParseAddrPort("127.0.0.1:1")

	prober := Prober{PerAttemptTimeout: 50 * time.Millisecond}

	result, err := prober.ProbeFirstHealthy(context.Background(), []netip.AddrPort{unreachable, endpoint}, "health")
	if err != nil {
		t.Fatalf("ProbeFirstHealthy returned error: %v", err)
	}

	if result.Endpoint != endpoint {
		t.Fatalf("endpoint = %s, want %s", result.Endpoint, endpoint)
	}
	if result.Reply != "ok" {
		t.Fatalf("reply = %q, want ok", result.Reply)
	}
}

func TestProbeFirstHealthyTimesOutSlowEndpoint(t *testing.T) {
	listener := startProbeServer(t, 100*time.Millisecond, "ok\n")
	defer listener.Close()

	endpoint := mustAddrPort(t, listener.Addr().String())
	prober := Prober{PerAttemptTimeout: 20 * time.Millisecond}

	_, err := prober.ProbeFirstHealthy(context.Background(), []netip.AddrPort{endpoint}, "health")
	if err == nil {
		t.Fatalf("ProbeFirstHealthy error = nil, want timeout")
	}
	if !Timeout(err) {
		t.Fatalf("ProbeFirstHealthy error = %v, want timeout", err)
	}
}

func TestProbeFirstHealthyRejectsEmptyEndpointList(t *testing.T) {
	prober := Prober{}

	_, err := prober.ProbeFirstHealthy(context.Background(), nil, "health")
	if err != ErrNoEndpoints {
		t.Fatalf("ProbeFirstHealthy error = %v, want ErrNoEndpoints", err)
	}
}

func startProbeServer(t *testing.T, delay time.Duration, reply string) net.Listener {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen returned error: %v", err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			go func(conn net.Conn) {
				defer conn.Close()

				if delay > 0 {
					time.Sleep(delay)
				}

				_, _ = conn.Write([]byte(reply))
			}(conn)
		}
	}()

	return listener
}

func mustAddrPort(t *testing.T, value string) netip.AddrPort {
	t.Helper()

	addr, err := ParseEndpoint(value)
	if err != nil {
		t.Fatalf("ParseEndpoint returned error: %v", err)
	}

	return addr
}
