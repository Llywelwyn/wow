package ui

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// Styles exposes the reusable primitives used to render wow output.
// Everything funnels through this struct so we can evolve the palette in one spot.
type Styles struct {
	Header    lipgloss.Style
	Icon      lipgloss.Style
	Key       lipgloss.Style
	Accent    lipgloss.Style
	Secondary lipgloss.Style
	Subtle    lipgloss.Style
	Label     lipgloss.Style
	Tag       lipgloss.Style
	Positive  lipgloss.Style
	Negative  lipgloss.Style
	Empty     lipgloss.Style
}

var (
	defaultStyles Styles
	once          sync.Once
)

// DefaultStyles returns the shared style set used across commands.
// It is initialised lazily so tests can tweak lipgloss globals before use.
func DefaultStyles() Styles {
	once.Do(func() {
		accent := lipgloss.AdaptiveColor{Light: "#005F87", Dark: "#5FD7FF"}
		secondary := lipgloss.AdaptiveColor{Light: "#5F5F87", Dark: "#9EA1E1"}
		subtle := lipgloss.AdaptiveColor{Light: "#6C6C6C", Dark: "#AAAAAA"}
		success := lipgloss.AdaptiveColor{Light: "#2B8A3E", Dark: "#81C995"}
		danger := lipgloss.AdaptiveColor{Light: "#C92A2A", Dark: "#FF6B6B"}

		defaultStyles = Styles{
			Header: lipgloss.NewStyle().
				Foreground(accent).
				Bold(true),
			Icon: lipgloss.NewStyle().
				Foreground(accent),
			Key: lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#212529", Dark: "#ECEFF4"}).
				Bold(true),
			Accent: lipgloss.NewStyle().
				Foreground(accent),
			Secondary: lipgloss.NewStyle().
				Foreground(secondary),
			Subtle: lipgloss.NewStyle().
				Foreground(subtle),
			Label: lipgloss.NewStyle().
				Foreground(subtle).
				Bold(true),
			Tag: lipgloss.NewStyle().
				Foreground(secondary).
				Bold(true).Underline(true),
			Positive: lipgloss.NewStyle().
				Foreground(success),
			Negative: lipgloss.NewStyle().
				Foreground(danger),
			Empty: lipgloss.NewStyle().
				Foreground(subtle).
				Italic(true),
		}
	})
	return defaultStyles
}
