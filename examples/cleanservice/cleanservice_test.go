package cleanservice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestServiceCreatePaymentMapsDTOAndEmitsAudit(t *testing.T) {
	store := &memoryStore{}
	auditCh := make(chan AuditEvent, 1)
	now := func() time.Time {
		return time.Date(2026, time.March, 22, 10, 30, 0, 0, time.UTC)
	}

	service := NewService(store, staticIDs{id: "pay_123"}, now, auditCh)

	resp, err := service.CreatePayment(context.Background(), CreatePaymentRequest{
		PartnerID: "partner_42",
		Amount:    1999,
		Currency:  "GBP",
	})
	if err != nil {
		t.Fatalf("CreatePayment returned error: %v", err)
	}

	if resp.ID != "pay_123" {
		t.Fatalf("response ID = %q, want pay_123", resp.ID)
	}
	if resp.CreatedAt != "2026-03-22T10:30:00Z" {
		t.Fatalf("response CreatedAt = %q, want RFC3339 value", resp.CreatedAt)
	}

	saved := store.Last()
	if saved.PartnerID != "partner_42" || saved.Amount != 1999 {
		t.Fatalf("saved payment = %+v, want persisted request values", saved)
	}

	select {
	case event := <-auditCh:
		if event.PaymentID != "pay_123" || event.EventType != "payment.created" {
			t.Fatalf("audit event = %+v, want created event", event)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("expected audit event")
	}
}

func TestHandlerRejectsInvalidDTO(t *testing.T) {
	service := NewService(&memoryStore{}, staticIDs{id: "pay_123"}, time.Now, make(chan AuditEvent, 1))
	handler := NewHandler(service)
	mux := http.NewServeMux()
	handler.Register(mux)

	reqBody := []byte(`{"partner_id":"partner_42","amount":0,"currency":"GBP","unexpected":true}`)
	req := httptestRequest(t, http.MethodPost, "/payments", reqBody)
	recorder := httptestRecorder()

	mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
}

func TestAppShutdownDrainsAuditAndStopsServer(t *testing.T) {
	store := &memoryStore{}
	sink := &recordingSink{}
	app := NewApp("127.0.0.1:0", Dependencies{
		Store:     store,
		IDs:       staticIDs{id: "pay_789"},
		AuditSink: sink,
		Now: func() time.Time {
			return time.Date(2026, time.March, 22, 12, 0, 0, 0, time.UTC)
		},
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen returned error: %v", err)
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- app.Serve(listener)
	}()

	client := &http.Client{Timeout: 2 * time.Second}
	payload, _ := json.Marshal(CreatePaymentRequest{
		PartnerID: "partner_77",
		Amount:    4500,
		Currency:  "GBP",
	})

	resp, err := client.Post("http://"+listener.Addr().String()+"/payments", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("POST returned error: %v", err)
	}
	resp.Body.Close()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}

	if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		t.Fatalf("Serve error = %v, want http.ErrServerClosed", err)
	}

	if got := sink.Events(); len(got) != 1 || got[0].PaymentID != "pay_789" {
		t.Fatalf("audit events = %+v, want one drained event", got)
	}

	_, err = client.Post("http://"+listener.Addr().String()+"/payments", "application/json", bytes.NewReader(payload))
	if err == nil {
		t.Fatalf("POST after shutdown error = nil, want connection failure")
	}
}

type memoryStore struct {
	mu       sync.Mutex
	payments []Payment
}

func (s *memoryStore) Save(_ context.Context, payment Payment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.payments = append(s.payments, payment)
	return nil
}

func (s *memoryStore) Last() Payment {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.payments[len(s.payments)-1]
}

type staticIDs struct {
	id string
}

func (s staticIDs) Next() string {
	return s.id
}

type recordingSink struct {
	mu     sync.Mutex
	events []AuditEvent
}

func (s *recordingSink) Write(_ context.Context, event AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
	return nil
}

func (s *recordingSink) Events() []AuditEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	cloned := make([]AuditEvent, len(s.events))
	copy(cloned, s.events)
	return cloned
}

type responseRecorder struct {
	header http.Header
	body   bytes.Buffer
	Code   int
}

func httptestRecorder() *responseRecorder {
	return &responseRecorder{header: make(http.Header)}
}

func (r *responseRecorder) Header() http.Header {
	return r.header
}

func (r *responseRecorder) Write(p []byte) (int, error) {
	return r.body.Write(p)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.Code = statusCode
}

func httptestRequest(t *testing.T, method, target string, body []byte) *http.Request {
	t.Helper()

	req, err := http.NewRequest(method, target, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("http.NewRequest returned error: %v", err)
	}
	return req
}
