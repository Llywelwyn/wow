package services

import (
	"bytes"
	"strings"
)

func detectType(payload []byte) string {
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
