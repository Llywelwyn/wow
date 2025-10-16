package model

import "time"

// Snippet represents metadata stored in the database.
type Snippet struct {
	Key         string
	Type        string
	Created     time.Time
	Modified    time.Time
	Description string
	Tags        string
}
