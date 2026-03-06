package dialbudget

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"strings"
	"time"
)

var (
	ErrNoEndpoints       = errors.New("at least one endpoint is required")
	ErrNoHealthyEndpoint = errors.New("no healthy endpoint responded")
)

type ProbeResult struct {
	Endpoint  netip.AddrPort
	Reply     string
	Latency   time.Duration
	ProbeSent string
}

type Prober struct {
	Dialer            net.Dialer
	PerAttemptTimeout time.Duration
}

func (p Prober) ProbeFirstHealthy(ctx context.Context, endpoints []netip.AddrPort, probe string) (ProbeResult, error) {
	if len(endpoints) == 0 {
		return ProbeResult{}, ErrNoEndpoints
	}

	attemptTimeout := p.PerAttemptTimeout
	if attemptTimeout <= 0 {
		attemptTimeout = 150 * time.Millisecond
	}

	var lastErr error

	for _, endpoint := range endpoints {
		attemptCtx, cancel := context.WithTimeout(ctx, attemptTimeout)
		start := time.Now()

		conn, err := p.Dialer.DialContext(attemptCtx, "tcp", endpoint.String())
		cancel()
		if err != nil {
			lastErr = err
			continue
		}

		deadline := time.Now().Add(attemptTimeout)
		if err := conn.SetDeadline(deadline); err != nil {
			_ = conn.Close()
			lastErr = err
			continue
		}

		reply, err := exchangeProbe(conn, probe)
		latency := time.Since(start)
		_ = conn.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if strings.TrimSpace(reply) != "ok" {
			lastErr = fmt.Errorf("endpoint %s replied %q", endpoint, strings.TrimSpace(reply))
			continue
		}

		return ProbeResult{
			Endpoint:  endpoint,
			Reply:     strings.TrimSpace(reply),
			Latency:   latency,
			ProbeSent: probe,
		}, nil
	}

	if ctx.Err() != nil {
		return ProbeResult{}, ctx.Err()
	}
	if lastErr == nil {
		lastErr = ErrNoHealthyEndpoint
	}

	return ProbeResult{}, lastErr
}

func exchangeProbe(conn net.Conn, probe string) (string, error) {
	if _, err := fmt.Fprintf(conn, "%s\n", probe); err != nil {
		return "", err
	}

	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", err
	}

	return reply, nil
}

func ParseEndpoint(addr string) (netip.AddrPort, error) {
	return netip.ParseAddrPort(addr)
}

func Timeout(err error) bool {
	return os.IsTimeout(err)
}
