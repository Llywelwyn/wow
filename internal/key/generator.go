package key

import (
	"errors"
	"fmt"
	"time"
)

// ExistsFunc reports whether a normalized key already exists.
type ExistsFunc func(key string) (bool, error)

// GenerateAuto produces a key in the auto/<epoch> namespace.
// It appends a numeric suffix when collisions occur.
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
