package core

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/llywelwyn/wow/internal/key"
	"github.com/llywelwyn/wow/internal/model"
	"github.com/llywelwyn/wow/internal/storage"
)

// ErrSnippetExists indicates a save attempted to overwrite an existing snippet.
var ErrSnippetExists = errors.New("snippet already exists")

// SaveRequest captures the inputs required to persist a snippet.
type SaveRequest struct {
	Key         string
	Description string
	Tags        []string
	Reader      io.Reader
}

// SaveResult returns the persisted key, metadata, and captured contents.
type SaveResult struct {
	Key      string
	Metadata model.Metadata
	Contents []byte
}

// Saver coordinates saving snippet content and metadata.
type Saver struct {
	BaseDir string
	DB      *sql.DB
	Now     func() time.Time
}

// Save writes the snippet to disk and stores metadata, generating an auto key when absent.
func (s *Saver) Save(ctx context.Context, req SaveRequest) (SaveResult, error) {
	if s.DB == nil || s.Now == nil {
		return SaveResult{}, errors.New("saver misconfigured")
	}
	if req.Reader == nil {
		return SaveResult{}, errors.New("reader required")
	}

	payload, err := io.ReadAll(req.Reader)
	if err != nil {
		return SaveResult{}, fmt.Errorf("read input: %w", err)
	}

	if len(payload) == 0 {
		return SaveResult{}, errors.New("snippet content is empty")
	}

	contentType := detectType(payload)
	now := s.Now()

	resolvedKey, err := s.resolveKey(ctx, req.Key, now)
	if err != nil {
		return SaveResult{}, err
	}

	path, err := key.ResolvePath(s.BaseDir, resolvedKey)
	if err != nil {
		return SaveResult{}, err
	}

	exists, err := storage.Exists(path)
	if err != nil {
		return SaveResult{}, err
	}
	if exists {
		return SaveResult{}, ErrSnippetExists
	}

	if err := storage.Save(path, bytes.NewReader(payload)); err != nil {
		return SaveResult{}, err
	}

	meta := model.Metadata{
		Key:         resolvedKey,
		Type:        contentType,
		Created:     now,
		Modified:    now,
		Description: req.Description,
		Tags:        normalizeTags(req.Tags),
	}

	if err := storage.InsertMetadata(ctx, s.DB, meta); err != nil {
		_ = storage.Delete(path)
		if errors.Is(err, storage.ErrMetadataDuplicate) {
			return SaveResult{}, ErrSnippetExists
		}
		return SaveResult{}, err
	}

	return SaveResult{
		Key:      resolvedKey,
		Metadata: meta,
		Contents: payload,
	}, nil
}

func (s *Saver) resolveKey(ctx context.Context, rawKey string, now time.Time) (string, error) {
	if strings.TrimSpace(rawKey) != "" {
		return key.Normalize(rawKey)
	}

	cb := func(candidate string) (bool, error) {
		path, err := key.ResolvePath(s.BaseDir, candidate)
		if err != nil {
			return false, err
		}
		return storage.Exists(path)
	}

	autoKey, err := key.GenerateAuto(now, cb)
	if err != nil {
		return "", err
	}
	return autoKey, nil
}

func detectType(payload []byte) string {
	firstLine := firstLine(payload)
	if isURL(firstLine) {
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

func normalizeTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}

	var normalized []string
	seen := make(map[string]struct{})
	for _, t := range tags {
		t = strings.TrimSpace(strings.ToLower(t))
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		normalized = append(normalized, t)
	}
	return strings.Join(normalized, ",")
}
