package cloudy

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
)

func IsZeroDate(dt strfmt.DateTime) bool {
	return time.Time(dt).IsZero()
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func NewByteReadCloser(data []byte) io.ReadCloser {
	return nopCloser{bytes.NewBuffer(data)}
}

type ByteCounter struct {
	total  int64
	reader io.ReadCloser
}

func NewByteCounter(stream io.ReadCloser) *ByteCounter {
	return &ByteCounter{
		reader: stream,
	}
}

func (bc *ByteCounter) Read(buf []byte) (n int, err error) {
	read, err := bc.reader.Read(buf)
	bc.total += int64(read)
	return read, err
}

func (bc *ByteCounter) Close() error {
	return bc.reader.Close()
}

func (bc *ByteCounter) Total() int64 {
	return bc.total
}

func DeferableClose(ctx context.Context, closeme io.Closer) {
	if err := closeme.Close(); err != nil {
		Error(ctx, "Error closing: %s\n", err)
	}
}

func IsMap(suspect interface{}) bool {
	mp := suspect.(map[string]interface{})
	return mp != nil
}

func HasPrefixOverlap(s string, prefix string) bool {
	if strings.HasPrefix(s, prefix) {
		return true
	}
	// If the input string is smaller than the prefix, it cannot have a the full prefix
	if len(prefix) > len(s) {
		return strings.HasPrefix(prefix, s)
	}
	return false
}
