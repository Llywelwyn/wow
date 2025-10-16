package model

import "time"

// Metadata captures descriptive fields persisted alongside snippet contents.
type Metadata struct {
	Key         string
	Type        string
	Created     time.Time
	Modified    time.Time
	Description string
	Tags        string
}
