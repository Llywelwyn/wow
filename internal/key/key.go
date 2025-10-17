// key manages wow keys.
// It handles the normalisation and validation of raw input keys,
// and it resolves normalised keys to their absolute filepaths.
package key

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

var (
	ErrEmpty      = errors.New("key must not be empty")
	ErrAbsolute   = errors.New("key must be relative")
	ErrTraversal  = errors.New("key cannot traverse parent directories")
	ErrBadSegment = errors.New("key segment invalid")
	ErrBadRune    = errors.New("key contains an unsupported character")
)

// ResolvePath converts a normalized key into an absolute path under the provided base directory.
// It ensures the resolved path does not escape the base directory.
func ResolvePath(baseDir, key string) (string, error) {
	normalized, err := Normalize(key)
	if err != nil {
		return "", err
	}

	full := filepath.Join(baseDir, filepath.FromSlash(normalized))
	clean := filepath.Clean(full)

	rel, err := filepath.Rel(baseDir, clean)
	if err != nil {
		return "", fmt.Errorf("resolve key %q: %w", key, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", ErrTraversal
	}

	return clean, nil
}

// Normalize trims and validates the provided key.
// It returns the normalized representation
// or an error if the key is invalid.
func Normalize(raw string) (string, error) {
	key := strings.TrimSpace(raw)
	if key == "" {
		return "", ErrEmpty
	}

	if strings.HasPrefix(key, "/") {
		return "", ErrAbsolute
	}

	if strings.Contains(key, "//") {
		return "", fmt.Errorf("%w: empty segment", ErrBadSegment)
	}

	segments := strings.SplitSeq(key, "/")
	for seg := range segments {
		if err := validateSegment(seg); err != nil {
			return "", err
		}
	}

	return key, nil
}

// validateSegment checks if a key segment is valid.
//
// It returns one of the following errors:
//   - ErrBadSegment 	if the segment is empty.
//   - ErrTraversal 	if the segment attempts to traverse (e.g. ".", "..").
//   - ErrBadRune 		if the segment contains disallowed characters.
func validateSegment(seg string) error {
	if seg == "" {
		return fmt.Errorf("%w: empty segment", ErrBadSegment)
	}
	if seg == "." || seg == ".." {
		return ErrTraversal
	}

	for _, r := range seg {
		if !isAllowed(r) {
			return fmt.Errorf("%w: %q", ErrBadRune, r)
		}
	}

	return nil
}

// isAllowed reports whether the rune is permitted in a key segment.
func isAllowed(r rune) bool {
	switch {
	case r == '-', r == '_', r == '.':
		return true
	case r == '/':
		return false
	case unicode.IsDigit(r):
		return true
	case unicode.IsLetter(r):
		return r <= unicode.MaxASCII
	default:
		return false
	}
}
