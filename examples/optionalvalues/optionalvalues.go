package optionalvalues

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

type Optional[T any] struct {
	Value T
	Valid bool
}

func Opt[T any](v T) Optional[T] {
	return Optional[T]{
		Value: v,
		Valid: true,
	}
}

func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if !o.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(o.Value)
}

type Field[T any] struct {
	Value T
	Set   bool
	Valid bool
}

func Some[T any](v T) Field[T] {
	return Field[T]{
		Value: v,
		Set:   true,
		Valid: true,
	}
}

func Null[T any]() Field[T] {
	return Field[T]{
		Set:   true,
		Valid: false,
	}
}

func (f *Field[T]) UnmarshalJSON(data []byte) error {
	f.Set = true

	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		var zero T
		f.Value = zero
		f.Valid = false
		return nil
	}

	if err := json.Unmarshal(trimmed, &f.Value); err != nil {
		return err
	}

	f.Valid = true
	return nil
}

func (f Field[T]) MarshalJSON() ([]byte, error) {
	if !f.Set || !f.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(f.Value)
}

var (
	ErrNullDisplayName         = errors.New("display_name must not be null")
	ErrEmptyDisplayName        = errors.New("display_name must not be empty")
	ErrNullSettlementDelayDays = errors.New("settlement_delay_days must not be null")
	ErrNegativeSettlementDelay = errors.New("settlement_delay_days must be >= 0")
)

type Transaction struct {
	ID     string `json:"id"`
	Amount int    `json:"amount"`
}

type TransactionPage struct {
	Data       []Transaction    `json:"data"`
	HasMore    bool             `json:"has_more"`
	NextCursor Optional[string] `json:"next_cursor"`
	TotalCount Optional[int]    `json:"total_count"`
}

type Partner struct {
	ID                  string
	DisplayName         string
	SettlementDelayDays int
	ExternalRef         Optional[string]
}

type UpdatePartnerRequest struct {
	DisplayName         Field[string] `json:"display_name"`
	SettlementDelayDays Field[int]    `json:"settlement_delay_days"`
	ExternalRef         Field[string] `json:"external_ref"`
}

func DecodeUpdatePartnerRequest(body []byte) (UpdatePartnerRequest, error) {
	var req UpdatePartnerRequest

	if err := json.Unmarshal(body, &req); err != nil {
		return UpdatePartnerRequest{}, fmt.Errorf("decode update request: %w", err)
	}

	return req, nil
}

func ApplyPartnerPatch(dst *Partner, req UpdatePartnerRequest) error {
	if req.DisplayName.Set {
		if !req.DisplayName.Valid {
			return ErrNullDisplayName
		}
		if req.DisplayName.Value == "" {
			return ErrEmptyDisplayName
		}

		dst.DisplayName = req.DisplayName.Value
	}

	if req.SettlementDelayDays.Set {
		if !req.SettlementDelayDays.Valid {
			return ErrNullSettlementDelayDays
		}
		if req.SettlementDelayDays.Value < 0 {
			return ErrNegativeSettlementDelay
		}

		dst.SettlementDelayDays = req.SettlementDelayDays.Value
	}

	if req.ExternalRef.Set {
		if !req.ExternalRef.Valid {
			dst.ExternalRef = Optional[string]{}
		} else {
			dst.ExternalRef = Opt(req.ExternalRef.Value)
		}
	}

	return nil
}
