package key

import (
	"errors"
	"testing"
	"time"
)

func TestGenerateAutoProducesBaseKeyWhenAvailable(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)

	calls := 0
	mockExists := func(k string) (bool, error) {
		calls++
		if k == "auto/1700000000" {
			return false, nil
		}
		return false, errors.New("unexpected key")
	}

	key, err := GenerateAuto(now, mockExists)
	if err != nil {
		t.Fatalf("GenerateAuto error = %v", err)
	}
	if key != "auto/1700000000" {
		t.Fatalf("key = %q, want auto/1700000000", key)
	}
	if calls != 1 {
		t.Fatalf("exists called %d times, want 1", calls)
	}
}

func TestGenerateAutoAddsSuffixOnCollision(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)

	calls := 0
	mockExists := func(k string) (bool, error) {
		calls++
		switch k {
		case "auto/1700000000":
			return true, nil
		case "auto/1700000000-1":
			return true, nil
		case "auto/1700000000-2":
			return false, nil
		default:
			return false, errors.New("unexpected key")
		}
	}

	key, err := GenerateAuto(now, mockExists)
	if err != nil {
		t.Fatalf("GenerateAuto error = %v", err)
	}
	if key != "auto/1700000000-2" {
		t.Fatalf("key = %q, want auto/1700000000-2", key)
	}
	if calls != 3 {
		t.Fatalf("exists called %d times, want 3", calls)
	}
}

func TestGenerateAutoForwardsErrors(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	mockExists := func(string) (bool, error) {
		return false, errors.New("boom")
	}

	_, err := GenerateAuto(now, mockExists)
	if err == nil || err.Error() != `check existing key "auto/1700000000": boom` {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestGenerateAutoRequiresCallback(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	if _, err := GenerateAuto(now, nil); err == nil {
		t.Fatalf("expected error when exists callback missing")
	}
}
