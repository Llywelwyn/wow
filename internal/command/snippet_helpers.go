package command

import (
	"bytes"
	"slices"
	"strings"
)

// detectSnippetType inspects the payload to infer a snippet type.
func detectSnippetType(payload []byte) string {
	first := firstLine(payload)
	if isURL(first) {
		return "url"
	}
	return "text"
}

func firstLine(payload []byte) string {
	idx := bytes.IndexByte(payload, '\n')
	if idx == -1 {
		return strings.TrimSpace(string(payload))
	}
	return strings.TrimSpace(string(payload[:idx]))
}

func isURL(line string) bool {
	lower := strings.ToLower(line)
	return strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "ftp://") ||
		strings.HasPrefix(lower, "file://")
}

// mergeTags returns the canonical tag string after applying additions/removals.
func mergeTags(existing string, add []string, remove []string) string {
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

// parseTags returns unique, normalized tags from a CSV string.
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

// normalizeTag trims and lowercases a tag value.
func normalizeTag(tag string) string {
	return strings.TrimSpace(strings.ToLower(tag))
}

// diffTags returns the items in a that are not present in b.
func diffTags(a, b []string) []string {
	if len(a) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(b))
	for _, tag := range b {
		set[tag] = struct{}{}
	}
	var out []string
	for _, tag := range a {
		if _, ok := set[tag]; ok {
			continue
		}
		out = append(out, tag)
	}
	return out
}

// firstNonEmptyLine returns the first non-empty trimmed line from data.
func firstNonEmptyLine(data []byte) string {
	lines := strings.Split(string(data), "\n")
	for _, l := range lines {
		if trimmed := strings.TrimSpace(l); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
