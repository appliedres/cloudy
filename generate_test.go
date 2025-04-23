package cloudy

import (
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePassword(t *testing.T) {
	p := GeneratePassword(8, 2, 2, 2)

	assert.Equal(t, len(p), 8)
	assert.Regexp(t, regexp.MustCompile(".*[0-9].*[0-9].*"), p, "Expecting 2 digits")
	assert.Regexp(t, regexp.MustCompile(".*[A-Z].*[A-Z].*"), p, "Expecting 2 uppercase")
	assert.Regexp(t, regexp.MustCompile(".*[!@#$%&*].*[!@#$%&*].*"), p, "Expecting 2 special")

	pv := IsValidPassword(p)
	assert.True(t, pv, p)

	pv = IsValidPasswordNoSpecial(p)
	assert.False(t, pv, p)

	p = GeneratePasswordNoSpecial(8, 2, 2)

	assert.Equal(t, len(p), 8)
	assert.Regexp(t, regexp.MustCompile(".*[0-9].*[0-9].*"), p, "Expecting 2 digits")
	assert.Regexp(t, regexp.MustCompile(".*[A-Z].*[A-Z].*"), p, "Expecting 2 uppercase")
	assert.Regexp(t, regexp.MustCompile(".*[!@#$%&*]{0}.*"), p, "Expecting 0 special")

	pv = IsValidPasswordNoSpecial(p)
	assert.True(t, pv, p)

	pv = IsValidPassword(p)
	assert.False(t, pv, p)


	pv = IsValidPasswordNoSpecial("testpassword123456")
	assert.False(t, pv, p)
}

func TestGeneratedD(t *testing.T) {
	id1 := GenerateId("ID", 10)
	id2 := GenerateId("ID", 10)

	assert.Equal(t, len(id1), 10)
	assert.Equal(t, len(id2), 10)
	assert.NotEqual(t, id1, id2)
}

func TestHashId(t *testing.T) {
	id := HashId("id", "1234124", "asdfasfdas", "gfsdgfsdg")
	assert.NotNil(t, id)
}

func TestGenerateVMIDFromPrefix_Success(t *testing.T) {
	// Reset counter (for consistency, though timestamp part isn't asserted here)
	atomic.StoreUint32(&timestampGenCounter, 0)

	id, err := GenerateVMIDFromPrefix("uvm")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.HasPrefix(id, "uvm-") {
		t.Errorf("expected prefix \"uvm-\", got %q", id)
	}

	// Should be exactly prefix length + 10
	expectedLen := len("uvm-") + len("0123456789")
	if len(id) != expectedLen {
		t.Errorf("expected length %d, got %d", expectedLen, len(id))
	}
}

func TestGenerateVMIDFromPrefix_PrefixTooLong(t *testing.T) {
	// A prefix longer than 4 chars will overflow: 5 + 1 + 10 = 16 > 15
	_, err := GenerateVMIDFromPrefix("toolong")
	if err == nil {
		t.Fatal("expected error for long prefix, got nil")
	}
}

func TestGenerateUVMID(t *testing.T) {
	atomic.StoreUint32(&timestampGenCounter, 0)
	id := GenerateUVMID()

	if !strings.HasPrefix(id, "uvm-") {
		t.Errorf("GenerateUVMID: expected prefix \"uvm-\", got %q", id)
	}

	if len(id) != len("uvm-")+10 {
		t.Errorf("GenerateUVMID: expected length %d, got %d", len("uvm-")+10, len(id))
	}
}

func TestGenerateSHVMID(t *testing.T) {
	atomic.StoreUint32(&timestampGenCounter, 0)
	id := GenerateSHVMID()

	if !strings.HasPrefix(id, "shvm-") {
		t.Errorf("GenerateSHVMID: expected prefix \"shvm-\", got %q", id)
	}

	if len(id) != len("shvm-")+10 {
		t.Errorf("GenerateSHVMID: expected length %d, got %d", len("shvm-")+10, len(id))
	}
}


func TestGenerateTimestampID_FormatAndLength(t *testing.T) {
	// Reset counter for deterministic behavior
	atomic.StoreUint32(&timestampGenCounter, 0)

	// Use a fixed time: 2021-12-01T00:00:00Z
	t0 := time.Date(2021, time.December, 1, 0, 0, 0, 0, time.UTC)
	id := GenerateTimestampID(t0)

	// Should be exactly 10 characters
	if len(id) != 10 {
		t.Errorf("expected id length 10, got %d", len(id))
	}

	// Verify the timestamp prefix decodes correctly
	prefix := id[:9]
	parsedMs, err := strconv.ParseInt(prefix, 36, 64)
	if err != nil {
		t.Fatalf("failed to parse prefix %q: %v", prefix, err)
	}
	expectedMs := t0.UnixNano() / 1e6
	if parsedMs != expectedMs {
		t.Errorf("expected timestamp %d, got %d", expectedMs, parsedMs)
	}

	// Verify counter suffix: first call should be 1
	suffix := id[9:]
	if suffix != "1" {
		t.Errorf("expected suffix '1', got %q", suffix)
	}
}

func TestGenerateTimestampID_UniquenessAndWrap(t *testing.T) {
	// Set counter near wrap boundary
	atomic.StoreUint32(&timestampGenCounter, timestampGenCounterMod-1)
	// Fixed time
	t0 := time.Date(2021, time.December, 1, 0, 0, 0, 0, time.UTC)

	id1 := GenerateTimestampID(t0)
	id2 := GenerateTimestampID(t0)

	if id1 == id2 {
		t.Error("expected unique IDs for consecutive calls, got identical IDs")
	}

	// After wrap: first suffix = 0, next = 1
	ex1 := "0"
	ex2 := "1"
	if id1[9:] != ex1 {
		t.Errorf("expected first suffix %q, got %q", ex1, id1[9:])
	}
	if id2[9:] != ex2 {
		t.Errorf("expected second suffix %q, got %q", ex2, id2[9:])
	}
}

func TestDecodeTimestampID_Success(t *testing.T) {
	atomic.StoreUint32(&timestampGenCounter, 0)
	// Use a time with millisecond precision
	t0 := time.Date(2022, time.April, 23, 12, 34, 56, 789000000, time.UTC)
	id := GenerateTimestampID(t0)

	decoded, err := DecodeTimestampID(id)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Decode returns time at millisecond resolution
	expected := time.UnixMilli(t0.UnixNano() / 1e6)
	if !decoded.Equal(expected) {
		t.Errorf("expected decoded time %v, got %v", expected, decoded)
	}
}

func TestDecodeTimestampID_InvalidLength(t *testing.T) {
	_, err := DecodeTimestampID("short")
	if err == nil {
		t.Error("expected error for invalid id length, got nil")
	}
}

func TestDecodeTimestampID_InvalidFormat(t *testing.T) {
	// 9 non-base36 chars + valid suffix
	id := "!!!!!!!!!!"
	_, err := DecodeTimestampID(id)
	if err == nil {
		t.Error("expected error for invalid base36 prefix, got nil")
	}
}

func TestGenerateTimestampIDNow_Length(t *testing.T) {
	// Simply verify it returns a string of the correct length
	id := GenerateTimestampIDNow()
	if len(id) != 10 {
		t.Errorf("GenerateTimestampIDNow returned id of length %d, want 10", len(id))
	}
}
