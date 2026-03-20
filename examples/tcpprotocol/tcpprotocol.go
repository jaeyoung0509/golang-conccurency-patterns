package tcpprotocol

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"os"
	"time"
)

const DefaultMaxFrameSize = 1 << 20

var ErrFrameTooLarge = errors.New("frame exceeds configured maximum")

type SessionConfig struct {
	MaxFrameSize int
	Deadline     time.Duration
}

func ReadFrame(r io.Reader, maxFrameSize int) ([]byte, error) {
	if maxFrameSize <= 0 {
		maxFrameSize = DefaultMaxFrameSize
	}

	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, err
	}

	size := int(binary.BigEndian.Uint32(header[:]))
	if size > maxFrameSize {
		return nil, ErrFrameTooLarge
	}

	payload := make([]byte, size)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func WriteFrame(w io.Writer, payload []byte) error {
	frame := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(frame[:4], uint32(len(payload)))
	copy(frame[4:], payload)

	return writeAll(w, frame)
}

func Exchange(conn net.Conn, request []byte, cfg SessionConfig) ([]byte, error) {
	if cfg.Deadline > 0 {
		if err := conn.SetDeadline(time.Now().Add(cfg.Deadline)); err != nil {
			return nil, err
		}
	}

	if err := WriteFrame(conn, request); err != nil {
		return nil, err
	}

	return ReadFrame(conn, cfg.MaxFrameSize)
}

func ServeOne(conn net.Conn, cfg SessionConfig, handler func([]byte) ([]byte, error)) error {
	if cfg.Deadline > 0 {
		if err := conn.SetDeadline(time.Now().Add(cfg.Deadline)); err != nil {
			return err
		}
	}

	request, err := ReadFrame(conn, cfg.MaxFrameSize)
	if err != nil {
		return err
	}

	response, err := handler(request)
	if err != nil {
		return err
	}

	return WriteFrame(conn, response)
}

func Timeout(err error) bool {
	return os.IsTimeout(err)
}

func writeAll(w io.Writer, data []byte) error {
	for len(data) > 0 {
		written, err := w.Write(data)
		if err != nil {
			return err
		}
		if written == 0 {
			return io.ErrShortWrite
		}
		data = data[written:]
	}
	return nil
}
