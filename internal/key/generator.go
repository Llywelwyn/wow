package key

import (
	"errors"
	"fmt"
)

// ExistsFunc reports whether a normalized key already exists.
type ExistsFunc func(key string) (bool, error)

// NextID produces a non-taken, numeric key.
func NextID(exists ExistsFunc) (string, error) {
	if exists == nil {
		return "", errors.New("exists callback is required")
	}
	id := 0
	for {
		key := idstr(id)
		hasKey, err := exists(key)
		if err != nil {
			return "", fmt.Errorf("check existing key: %q %w", key, err)
		}
		if !hasKey {
			return key, nil
		}
		id++
	}
}

func idstr(id int) string {
	return fmt.Sprintf("%d", id)
}
