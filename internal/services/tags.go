package services

import (
	"slices"
	"strings"
)

// MergeTags returns the canonical representation after applying additions/removals.
func MergeTags(existing string, add []string, remove []string) string {
	current := parseTags(existing)

	for _, tag := range add {
		tag = normalizeTag(tag)
		if tag == "" || slices.Contains(current, tag) {
			continue
		}
		current = append(current, tag)
	}

	if len(remove) > 0 {
		toRemove := make(map[string]struct{}, len(remove))
		for _, tag := range remove {
			tag = normalizeTag(tag)
			if tag == "" {
				continue
			}
			toRemove[tag] = struct{}{}
		}
		if len(toRemove) > 0 {
			var filtered []string
			for _, tag := range current {
				if _, drop := toRemove[tag]; !drop {
					filtered = append(filtered, tag)
				}
			}
			current = filtered
		}
	}

	if len(current) == 0 {
		return ""
	}
	return strings.Join(current, ",")
}

func parseTags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var result []string
	seen := make(map[string]struct{})
	for _, tag := range parts {
		tag = normalizeTag(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	return result
}

func normalizeTag(tag string) string {
	return strings.TrimSpace(strings.ToLower(tag))
}
