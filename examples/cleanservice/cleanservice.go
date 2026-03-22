package cleanservice

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	ErrPartnerIDRequired     = errors.New("partner_id is required")
	ErrAmountMustBeAboveZero = errors.New("amount must be above zero")
)

type CreatePaymentRequest struct {
	PartnerID string `json:"partner_id"`
	Amount    int    `json:"amount"`
	Currency  string `json:"currency"`
}

type PaymentResponse struct {
	ID        string `json:"id"`
	PartnerID string `json:"partner_id"`
	Amount    int    `json:"amount"`
	Currency  string `json:"currency"`
	CreatedAt string `json:"created_at"`
}

type Payment struct {
	ID        string
	PartnerID string
	Amount    int
	Currency  string
	CreatedAt time.Time
}

type AuditEvent struct {
	PaymentID string
	PartnerID string
	EventType string
}

type PaymentStore interface {
	Save(context.Context, Payment) error
}

type IDGenerator interface {
	Next() string
}

type AuditSink interface {
	Write(context.Context, AuditEvent) error
}

type Service struct {
	store PaymentStore
	ids   IDGenerator
	now   func() time.Time
	audit chan<- AuditEvent
}

func NewService(store PaymentStore, ids IDGenerator, now func() time.Time, audit chan<- AuditEvent) Service {
	if now == nil {
		now = time.Now
	}

	return Service{
		store: store,
		ids:   ids,
		now:   now,
		audit: audit,
	}
}

func (s Service) CreatePayment(ctx context.Context, req CreatePaymentRequest) (PaymentResponse, error) {
	if req.PartnerID == "" {
		return PaymentResponse{}, ErrPartnerIDRequired
	}
	if req.Amount <= 0 {
		return PaymentResponse{}, ErrAmountMustBeAboveZero
	}

	payment := Payment{
		ID:        s.ids.Next(),
		PartnerID: req.PartnerID,
		Amount:    req.Amount,
		Currency:  req.Currency,
		CreatedAt: s.now().UTC(),
	}

	if err := s.store.Save(ctx, payment); err != nil {
		return PaymentResponse{}, err
	}

	select {
	case s.audit <- AuditEvent{
		PaymentID: payment.ID,
		PartnerID: payment.PartnerID,
		EventType: "payment.created",
	}:
	case <-ctx.Done():
		return PaymentResponse{}, ctx.Err()
	}

	return toPaymentResponse(payment), nil
}

type Handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return Handler{service: service}
}

func (h Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /payments", h.createPayment)
}

func (h Handler) createPayment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req CreatePaymentRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	resp, err := h.service.CreatePayment(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrPartnerIDRequired), errors.Is(err, ErrAmountMustBeAboveZero):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

type Dependencies struct {
	Store     PaymentStore
	IDs       IDGenerator
	AuditSink AuditSink
	Now       func() time.Time
}

type App struct {
	server    *http.Server
	auditCh   chan AuditEvent
	auditWg   sync.WaitGroup
	auditSink AuditSink
}

func NewApp(addr string, deps Dependencies) *App {
	auditCh := make(chan AuditEvent, 16)
	service := NewService(deps.Store, deps.IDs, deps.Now, auditCh)
	mux := http.NewServeMux()
	NewHandler(service).Register(mux)

	app := &App{
		server: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 2 * time.Second,
			IdleTimeout:       30 * time.Second,
		},
		auditCh:   auditCh,
		auditSink: deps.AuditSink,
	}

	app.auditWg.Add(1)
	go app.runAuditWorker()

	return app
}

func (a *App) Serve(listener net.Listener) error {
	return a.server.Serve(listener)
}

func (a *App) Shutdown(ctx context.Context) error {
	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}

	close(a.auditCh)

	done := make(chan struct{})
	go func() {
		defer close(done)
		a.auditWg.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *App) runAuditWorker() {
	defer a.auditWg.Done()

	for event := range a.auditCh {
		if a.auditSink == nil {
			continue
		}
		_ = a.auditSink.Write(context.Background(), event)
	}
}

func toPaymentResponse(payment Payment) PaymentResponse {
	return PaymentResponse{
		ID:        payment.ID,
		PartnerID: payment.PartnerID,
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		CreatedAt: payment.CreatedAt.Format(time.RFC3339),
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
