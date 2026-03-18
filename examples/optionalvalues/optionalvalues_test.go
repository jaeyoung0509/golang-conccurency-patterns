package optionalvalues

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestTransactionPageMarshalsExplicitNulls(t *testing.T) {
	page := TransactionPage{
		Data: []Transaction{
			{ID: "TRX-1999", Amount: 3000},
		},
		HasMore: false,
	}

	raw, err := json.Marshal(page)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	got := string(raw)
	if !strings.Contains(got, `"next_cursor":null`) {
		t.Fatalf("page JSON missing explicit null next_cursor: %s", got)
	}
	if !strings.Contains(got, `"total_count":null`) {
		t.Fatalf("page JSON missing explicit null total_count: %s", got)
	}
}

func TestTransactionPageMarshalsOptionalValues(t *testing.T) {
	page := TransactionPage{
		Data: []Transaction{
			{ID: "TRX-1001", Amount: 5000},
			{ID: "TRX-1002", Amount: 12000},
		},
		HasMore:    true,
		NextCursor: Opt("TRX-1002"),
		TotalCount: Opt(15000),
	}

	raw, err := json.Marshal(page)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	got := string(raw)
	if !strings.Contains(got, `"next_cursor":"TRX-1002"`) {
		t.Fatalf("page JSON missing next_cursor value: %s", got)
	}
	if !strings.Contains(got, `"total_count":15000`) {
		t.Fatalf("page JSON missing total_count value: %s", got)
	}
}

func TestFieldDistinguishesAbsentNullAndValue(t *testing.T) {
	req, err := DecodeUpdatePartnerRequest([]byte(`{
		"display_name": "Acme Europe",
		"external_ref": null
	}`))
	if err != nil {
		t.Fatalf("DecodeUpdatePartnerRequest returned error: %v", err)
	}

	if !req.DisplayName.Set || !req.DisplayName.Valid || req.DisplayName.Value != "Acme Europe" {
		t.Fatalf("display_name decoded incorrectly: %+v", req.DisplayName)
	}
	if !req.ExternalRef.Set || req.ExternalRef.Valid {
		t.Fatalf("external_ref should be explicit null: %+v", req.ExternalRef)
	}
	if req.SettlementDelayDays.Set {
		t.Fatalf("settlement_delay_days should remain absent: %+v", req.SettlementDelayDays)
	}
}

func TestApplyPartnerPatchClearsOptionalAndPreservesAbsent(t *testing.T) {
	partner := Partner{
		ID:                  "partner-77",
		DisplayName:         "Acme",
		SettlementDelayDays: 2,
		ExternalRef:         Opt("legacy-ref"),
	}

	req, err := DecodeUpdatePartnerRequest([]byte(`{
		"external_ref": null,
		"display_name": "Acme Payments"
	}`))
	if err != nil {
		t.Fatalf("DecodeUpdatePartnerRequest returned error: %v", err)
	}

	if err := ApplyPartnerPatch(&partner, req); err != nil {
		t.Fatalf("ApplyPartnerPatch returned error: %v", err)
	}

	if partner.DisplayName != "Acme Payments" {
		t.Fatalf("DisplayName = %q, want updated value", partner.DisplayName)
	}
	if partner.SettlementDelayDays != 2 {
		t.Fatalf("SettlementDelayDays = %d, want unchanged value", partner.SettlementDelayDays)
	}
	if partner.ExternalRef.Valid {
		t.Fatalf("ExternalRef should be cleared: %+v", partner.ExternalRef)
	}
}

func TestApplyPartnerPatchRejectsNullRequiredField(t *testing.T) {
	partner := Partner{
		ID:                  "partner-88",
		DisplayName:         "Northwind",
		SettlementDelayDays: 1,
	}

	req, err := DecodeUpdatePartnerRequest([]byte(`{
		"display_name": null
	}`))
	if err != nil {
		t.Fatalf("DecodeUpdatePartnerRequest returned error: %v", err)
	}

	err = ApplyPartnerPatch(&partner, req)
	if !errors.Is(err, ErrNullDisplayName) {
		t.Fatalf("ApplyPartnerPatch error = %v, want ErrNullDisplayName", err)
	}
}
