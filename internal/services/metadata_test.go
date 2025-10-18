package services

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/llywelwyn/wow/internal/storage"
)

func TestMetadataUpdateTags(t *testing.T) {
	base := t.TempDir()
	dbPath := filepath.Join(base, "meta.db")
	db, err := storage.InitMetaDB(dbPath)
	if err != nil {
		t.Fatalf("InitMetaDB error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	saver := &Saver{
		BaseDir: base,
		DB:      db,
		Now: func() time.Time {
			return time.Unix(1_700_000_000, 0)
		},
	}

	ctx := context.Background()
	if _, err := saver.Save(ctx, SaveRequest{Key: "go/foo", Reader: strings.NewReader("data"), Tags: []string{"foo"}}); err != nil {
		t.Fatalf("Save error = %v", err)
	}

	metaSvc := &Metadata{DB: db, Now: func() time.Time { return time.Unix(1_700_000_100, 0) }}
	result, err := metaSvc.UpdateTags(ctx, "go/foo", []string{"bar"}, []string{"foo"})
	if err != nil {
		t.Fatalf("UpdateTags error = %v", err)
	}

	if result.Metadata.Tags != "bar" {
		t.Fatalf("Tags = %q, want bar", result.Metadata.Tags)
	}
	if len(result.Added) != 1 || result.Added[0] != "bar" {
		t.Fatalf("Added = %v, want [bar]", result.Added)
	}
	if len(result.Removed) != 1 || result.Removed[0] != "foo" {
		t.Fatalf("Removed = %v, want [foo]", result.Removed)
	}
}
