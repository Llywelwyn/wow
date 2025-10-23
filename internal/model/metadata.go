package model

import (
	"strconv"
	"strings"
	"time"
)

// Metadata captures descriptive fields persisted alongside snippet contents.
type Metadata struct {
	Key         string
	Type        string
	Created     time.Time
	Modified    time.Time
	Description string
	Tags        string
}

func (m *Metadata) Formatted(name string) string {
	switch name {
	case "Date":
		return m.DateStr()
	case "Type":
		return m.TypeStr()
	case "Name":
		return m.NameStr()
	case "Tags":
		return m.TagsStr()
	case "Desc":
		return m.DescStr()
	default:
		return ""
	}
}

func (m *Metadata) TypeIcon() string {
	switch m.Type {
	case "url":
		return "url"
	default:
		return "txt"
	}
}

func (m *Metadata) TypeStr() string {
	return m.TypeIcon()
}

func (m *Metadata) DateStr() string {
	return m.Modified.UTC().Format("02 Jan 15:04")
}

func (m *Metadata) NameStr() string {
	return m.Key
}

func (m *Metadata) TagsStr() string {
	if len(m.Tags) == 0 {
		return m.EmptyStr()
	}
	// Split CSV and prefix with @.
	// one,two,three -> @one @two @three
	split := strings.Split(m.Tags, ",")
	if len(split) == 1 {
		return split[0]
	} else {
		return split[0] + "(+" + strconv.Itoa(len(split[1:])) + ")"
	}
}

func (m *Metadata) EmptyStr() string {
	return "--"
}

func (m *Metadata) DescStr() string {
	if m.Description == "" {
		return m.EmptyStr()
	}
	return m.Description
}
