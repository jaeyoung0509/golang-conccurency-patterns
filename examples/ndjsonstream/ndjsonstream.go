package ndjsonstream

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
)

const defaultBufferSize = 16 * 1024

var ErrInvalidPreviewLimit = errors.New("preview limit must be at least 1")

type ShipmentEvent struct {
	OrderID     string `json:"order_id"`
	Stage       string `json:"stage"`
	Carrier     string `json:"carrier"`
	WeightGrams int    `json:"weight_grams"`
}

func EncodeStream(ctx context.Context, w io.Writer, events <-chan ShipmentEvent) error {
	buffered := bufio.NewWriterSize(w, defaultBufferSize)
	encoder := json.NewEncoder(buffered)
	encoder.SetEscapeHTML(false)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-events:
			if !ok {
				return buffered.Flush()
			}
			if err := encoder.Encode(event); err != nil {
				return err
			}
		}
	}
}

func FilterStage(r io.Reader, stage string) ([]byte, error) {
	decoder := json.NewDecoder(bufio.NewReaderSize(r, defaultBufferSize))
	decoder.DisallowUnknownFields()

	var out bytes.Buffer
	buffered := bufio.NewWriterSize(&out, defaultBufferSize)
	encoder := json.NewEncoder(buffered)
	encoder.SetEscapeHTML(false)

	for {
		var event ShipmentEvent

		err := decoder.Decode(&event)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		if event.Stage != stage {
			continue
		}

		if err := encoder.Encode(event); err != nil {
			return nil, err
		}
	}

	if err := buffered.Flush(); err != nil {
		return nil, err
	}

	return bytes.Clone(out.Bytes()), nil
}

func CopyPreview(dst io.Writer, src io.Reader, limit int64) (int64, error) {
	if limit < 1 {
		return 0, ErrInvalidPreviewLimit
	}

	return io.Copy(dst, io.LimitReader(src, limit))
}
