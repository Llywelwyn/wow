package key

import (
	"errors"
	"fmt"
	"time"
)

// ExistsFunc reports whether a normalized key already exists.
type ExistsFunc func(key string) (bool, error)

// GenerateAuto produces a key in the auto/<epoch> namespace.
// GenerateAuto generates a key in the "auto/<unix_epoch>" namespace and appends
// a numeric suffix (e.g., "-1", "-2", ...) if collisions occur.
// The provided `now` value supplies the Unix timestamp used for the base key.
// The `exists` callback must be non-nil; if it is nil the function returns an
// error "exists callback is required". GenerateAuto calls `exists(key)` to test
// for collisions and returns the first key for which `exists` reports false.
// If `exists` returns an error, that error is returned wrapped with the
// context `check existing key "<key>": ...`.
func GenerateAuto(now time.Time, exists ExistsFunc) (string, error) {
	if exists == nil {
		return "", errors.New("exists callback is required")
	}

	base := fmt.Sprintf("auto/%d", now.Unix())
	key := base
	suffix := 1

	for {
		hasKey, err := exists(key)
		if err != nil {
			return "", fmt.Errorf("check existing key %q: %w", key, err)
		}

		if !hasKey {
			return key, nil
		}

		key = fmt.Sprintf("%s-%d", base, suffix)
		suffix++
	}
}