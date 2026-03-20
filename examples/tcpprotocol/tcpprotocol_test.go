package tcpprotocol

import (
	"bytes"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

func TestWriteFrameHandlesShortWrites(t *testing.T) {
	writer := &shortWriter{limit: 3}

	if err := WriteFrame(writer, []byte("payment-approved")); err != nil {
		t.Fatalf("WriteFrame returned error: %v", err)
	}

	want := append([]byte{0, 0, 0, 16}, []byte("payment-approved")...)
	if !bytes.Equal(writer.buf.Bytes(), want) {
		t.Fatalf("frame bytes = %v, want %v", writer.buf.Bytes(), want)
	}
}

func TestReadFrameHandlesFragmentedInput(t *testing.T) {
	var frame bytes.Buffer
	if err := WriteFrame(&frame, []byte("settlement-batch")); err != nil {
		t.Fatalf("WriteFrame returned error: %v", err)
	}

	reader := &chunkedReader{
		data:      frame.Bytes(),
		chunkSize: 2,
	}

	got, err := ReadFrame(reader, 64)
	if err != nil {
		t.Fatalf("ReadFrame returned error: %v", err)
	}
	if string(got) != "settlement-batch" {
		t.Fatalf("ReadFrame payload = %q, want settlement-batch", got)
	}
}

func TestReadFrameRejectsOversizedPayload(t *testing.T) {
	var frame bytes.Buffer
	if err := WriteFrame(&frame, []byte("oversized-payload")); err != nil {
		t.Fatalf("WriteFrame returned error: %v", err)
	}

	_, err := ReadFrame(bytes.NewReader(frame.Bytes()), 4)
	if !errors.Is(err, ErrFrameTooLarge) {
		t.Fatalf("ReadFrame error = %v, want ErrFrameTooLarge", err)
	}
}

func TestExchangeRoundTripOverNetPipe(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- ServeOne(serverConn, SessionConfig{MaxFrameSize: 256}, func(request []byte) ([]byte, error) {
			return bytes.ToUpper(request), nil
		})
	}()

	reply, err := Exchange(clientConn, []byte("partner-sync"), SessionConfig{MaxFrameSize: 256, Deadline: 50 * time.Millisecond})
	if err != nil {
		t.Fatalf("Exchange returned error: %v", err)
	}
	if string(reply) != "PARTNER-SYNC" {
		t.Fatalf("reply = %q, want PARTNER-SYNC", reply)
	}
	if err := <-serverErr; err != nil {
		t.Fatalf("ServeOne returned error: %v", err)
	}
}

func TestExchangeHonorsDeadline(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go func() {
		defer serverConn.Close()

		request, err := ReadFrame(serverConn, 256)
		if err != nil {
			return
		}
		if string(request) != "risk-check" {
			return
		}

		time.Sleep(80 * time.Millisecond)
		_ = WriteFrame(serverConn, []byte("approved"))
	}()

	_, err := Exchange(clientConn, []byte("risk-check"), SessionConfig{MaxFrameSize: 256, Deadline: 20 * time.Millisecond})
	if err == nil {
		t.Fatalf("Exchange error = nil, want timeout")
	}
	if !Timeout(err) {
		t.Fatalf("Exchange error = %v, want timeout", err)
	}
}

type shortWriter struct {
	buf   bytes.Buffer
	limit int
}

func (w *shortWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	n := len(p)
	if w.limit > 0 && n > w.limit {
		n = w.limit
	}
	return w.buf.Write(p[:n])
}

type chunkedReader struct {
	data      []byte
	offset    int
	chunkSize int
}

func (r *chunkedReader) Read(p []byte) (int, error) {
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}

	n := len(p)
	if r.chunkSize > 0 && n > r.chunkSize {
		n = r.chunkSize
	}
	remaining := len(r.data) - r.offset
	if n > remaining {
		n = remaining
	}

	copy(p[:n], r.data[r.offset:r.offset+n])
	r.offset += n
	return n, nil
}
