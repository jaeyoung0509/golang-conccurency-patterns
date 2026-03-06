package ndjsonstream

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"
)

func TestEncodeStreamWritesNDJSON(t *testing.T) {
	events := make(chan ShipmentEvent, 2)
	events <- ShipmentEvent{OrderID: "ord-1", Stage: "packed", Carrier: "dhl", WeightGrams: 1800}
	events <- ShipmentEvent{OrderID: "ord-2", Stage: "shipped", Carrier: "ups", WeightGrams: 2100}
	close(events)

	var out bytes.Buffer
	if err := EncodeStream(context.Background(), &out, events); err != nil {
		t.Fatalf("EncodeStream returned error: %v", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(out.Bytes()))

	var decoded []ShipmentEvent
	for {
		var event ShipmentEvent
		err := decoder.Decode(&event)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("Decode returned error: %v", err)
		}
		decoded = append(decoded, event)
	}

	if len(decoded) != 2 {
		t.Fatalf("decoded %d events, want 2", len(decoded))
	}
	if decoded[0].OrderID != "ord-1" || decoded[1].OrderID != "ord-2" {
		t.Fatalf("decoded events out of order: %#v", decoded)
	}
}

func TestFilterStageKeepsOnlyMatchingRecords(t *testing.T) {
	input := bytes.NewBufferString(
		`{"order_id":"ord-1","stage":"packed","carrier":"dhl","weight_grams":1800}` + "\n" +
			`{"order_id":"ord-2","stage":"shipped","carrier":"ups","weight_grams":2100}` + "\n" +
			`{"order_id":"ord-3","stage":"packed","carrier":"fedex","weight_grams":1600}` + "\n",
	)

	filtered, err := FilterStage(input, "packed")
	if err != nil {
		t.Fatalf("FilterStage returned error: %v", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(filtered))

	var decoded []ShipmentEvent
	for {
		var event ShipmentEvent
		err := decoder.Decode(&event)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("Decode returned error: %v", err)
		}
		decoded = append(decoded, event)
	}

	if len(decoded) != 2 {
		t.Fatalf("decoded %d events, want 2", len(decoded))
	}
	if decoded[0].OrderID != "ord-1" || decoded[1].OrderID != "ord-3" {
		t.Fatalf("unexpected filtered records: %#v", decoded)
	}
}

func TestFilterStageRejectsUnknownFields(t *testing.T) {
	input := bytes.NewBufferString(`{"order_id":"ord-1","stage":"packed","carrier":"dhl","weight_grams":1800,"unexpected":true}` + "\n")

	if _, err := FilterStage(input, "packed"); err == nil {
		t.Fatalf("FilterStage error = nil, want unknown-field error")
	}
}

func TestCopyPreviewLimitsBytes(t *testing.T) {
	var out bytes.Buffer

	written, err := CopyPreview(&out, bytes.NewBufferString("tracking:ord-100"), 8)
	if err != nil {
		t.Fatalf("CopyPreview returned error: %v", err)
	}

	if written != 8 {
		t.Fatalf("written = %d, want 8", written)
	}
	if out.String() != "tracking" {
		t.Fatalf("preview = %q, want %q", out.String(), "tracking")
	}
}
